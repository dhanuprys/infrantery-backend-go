package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/dhanuprys/infrantery-backend-go/internal/core/domain"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/port"
	"github.com/dhanuprys/infrantery-backend-go/pkg/compression"
	"github.com/dhanuprys/infrantery-backend-go/pkg/crypto"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// MaxBackupSize is the maximum allowed backup file size (100 MB).
	MaxBackupSize = 100 * 1024 * 1024

	// archiveHeaderSize = magic(5) + version(1) + nonce(12) + salt(32) = 50 bytes.
	archiveHeaderSize = 5 + 1 + crypto.NonceSize + crypto.SaltSize
)

var (
	ErrBackupTooLarge         = errors.New("backup file exceeds maximum allowed size")
	ErrBackupInvalidFormat    = errors.New("invalid backup file format")
	ErrBackupVersionMismatch  = errors.New("unsupported backup version")
	ErrBackupDecryptionFailed = errors.New("decryption failed: wrong password or corrupted file")
)

// BackupService handles project backup and restore operations.
type BackupService struct {
	projectService *ProjectService
	projectRepo    port.ProjectRepository
	memberRepo     port.ProjectMemberRepository
	noteRepo       port.NoteRepository
	diagramRepo    port.DiagramRepository
	nodeRepo       port.NodeRepository
	nodeVaultRepo  port.NodeVaultRepository
	argon2Params   *Argon2Params
}

// NewBackupService creates a new BackupService.
func NewBackupService(
	projectService *ProjectService,
	projectRepo port.ProjectRepository,
	memberRepo port.ProjectMemberRepository,
	noteRepo port.NoteRepository,
	diagramRepo port.DiagramRepository,
	nodeRepo port.NodeRepository,
	nodeVaultRepo port.NodeVaultRepository,
	argon2Params *Argon2Params,
) *BackupService {
	return &BackupService{
		projectService: projectService,
		projectRepo:    projectRepo,
		memberRepo:     memberRepo,
		noteRepo:       noteRepo,
		diagramRepo:    diagramRepo,
		nodeRepo:       nodeRepo,
		nodeVaultRepo:  nodeVaultRepo,
		argon2Params:   argon2Params,
	}
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// CreateBackup collects all project data, serializes, compresses, encrypts,
// and returns the archive as an io.Reader along with a suggested filename.
func (s *BackupService) CreateBackup(
	ctx context.Context,
	projectID, userID primitive.ObjectID,
	password string,
) (io.Reader, string, error) {
	// 1. Verify permission
	if err := s.projectService.HasPermission(ctx, projectID, userID, domain.PermissionManageProject); err != nil {
		return nil, "", err
	}

	member, err := s.memberRepo.FindByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		return nil, "", fmt.Errorf("fetching member for backup: %w", err)
	}

	// 2. Collect all data
	payload, err := s.collectProjectData(ctx, projectID, member)
	if err != nil {
		return nil, "", fmt.Errorf("collecting project data: %w", err)
	}

	// 3. Build the encrypted archive
	archive, err := s.buildArchive(payload, password)
	if err != nil {
		return nil, "", fmt.Errorf("building archive: %w", err)
	}

	filename := fmt.Sprintf("%s_%s.infbk",
		sanitizeFilename(payload.Project.Name),
		time.Now().Format("20060102_150405"),
	)

	return bytes.NewReader(archive), filename, nil
}

// RestoreBackup reads an encrypted backup, decrypts, decompresses, validates,
// and inserts all data as a new project. The restoring user becomes the owner.
func (s *BackupService) RestoreBackup(
	ctx context.Context,
	userID primitive.ObjectID,
	password string,
	backupReader io.Reader,
) (*domain.Project, error) {
	// 1. Read and validate size
	data, err := io.ReadAll(io.LimitReader(backupReader, MaxBackupSize+1))
	if err != nil {
		return nil, fmt.Errorf("reading backup file: %w", err)
	}
	if len(data) > MaxBackupSize {
		return nil, ErrBackupTooLarge
	}

	// 2. Parse archive → decrypt → decompress → unmarshal
	payload, err := s.parseArchive(data, password)
	if err != nil {
		return nil, err
	}

	// 3. Insert into database
	project, err := s.insertRestoredData(ctx, userID, payload)
	if err != nil {
		return nil, fmt.Errorf("inserting restored data: %w", err)
	}

	return project, nil
}

// ---------------------------------------------------------------------------
// Data Collection
// ---------------------------------------------------------------------------

func (s *BackupService) collectProjectData(
	ctx context.Context,
	projectID primitive.ObjectID,
	member *domain.ProjectMember,
) (*domain.BackupPayload, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("fetching project: %w", err)
	}

	diagrams, err := s.diagramRepo.FindAllByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("fetching diagrams: %w", err)
	}

	// Collect diagram IDs for bulk node fetch
	var nodes []*domain.Node
	if len(diagrams) > 0 {
		diagramIDs := make([]primitive.ObjectID, len(diagrams))
		for i, d := range diagrams {
			diagramIDs[i] = d.ID
		}

		nodes, err = s.nodeRepo.FindByDiagramIDs(ctx, diagramIDs)
		if err != nil {
			return nil, fmt.Errorf("fetching nodes: %w", err)
		}
	}

	vaults, err := s.nodeVaultRepo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("fetching vaults: %w", err)
	}

	notes, err := s.noteRepo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("fetching notes: %w", err)
	}

	return &domain.BackupPayload{
		Version:   domain.BackupVersion,
		CreatedAt: time.Now().UTC(),
		Project:   toProjectBackup(project),
		Member:    toMemberBackup(member),
		Diagrams:  toDiagramBackups(diagrams),
		Nodes:     toNodeBackups(nodes),
		Vaults:    toVaultBackups(vaults),
		Notes:     toNoteBackups(notes),
	}, nil
}

// ---------------------------------------------------------------------------
// Archive Building (serialize → compress → encrypt)
// ---------------------------------------------------------------------------

func (s *BackupService) buildArchive(payload *domain.BackupPayload, password string) ([]byte, error) {
	// 1. Serialize to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling payload: %w", err)
	}

	// 2. Compress
	compressed, err := compression.Compress(jsonData)
	if err != nil {
		return nil, fmt.Errorf("compressing payload: %w", err)
	}

	// 3. Derive encryption key
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("generating salt: %w", err)
	}

	key := crypto.DeriveBackupKey(password, domain.BackupPepper, salt, s.toCryptoParams())

	// 4. Encrypt
	nonce, ciphertext, err := crypto.Encrypt(compressed, key)
	if err != nil {
		return nil, fmt.Errorf("encrypting payload: %w", err)
	}

	// 5. Assemble archive: magic + version + nonce + salt + ciphertext
	var buf bytes.Buffer
	buf.Grow(archiveHeaderSize + len(ciphertext))
	buf.Write(domain.BackupMagic)
	buf.WriteByte(byte(domain.BackupVersion))
	buf.Write(nonce)
	buf.Write(salt)
	buf.Write(ciphertext)

	return buf.Bytes(), nil
}

// ---------------------------------------------------------------------------
// Archive Parsing (validate → decrypt → decompress → unmarshal)
// ---------------------------------------------------------------------------

func (s *BackupService) parseArchive(data []byte, password string) (*domain.BackupPayload, error) {
	if len(data) < archiveHeaderSize {
		return nil, ErrBackupInvalidFormat
	}

	// 1. Validate magic
	if !bytes.Equal(data[:5], domain.BackupMagic) {
		return nil, ErrBackupInvalidFormat
	}

	// 2. Validate version
	version := data[5]
	if int(version) != domain.BackupVersion {
		return nil, ErrBackupVersionMismatch
	}

	// 3. Extract nonce and salt
	offset := 6
	nonce := data[offset : offset+crypto.NonceSize]
	offset += crypto.NonceSize
	salt := data[offset : offset+crypto.SaltSize]
	offset += crypto.SaltSize
	ciphertext := data[offset:]

	// 4. Derive key and decrypt
	key := crypto.DeriveBackupKey(password, domain.BackupPepper, salt, s.toCryptoParams())

	compressed, err := crypto.Decrypt(ciphertext, key, nonce)
	if err != nil {
		return nil, ErrBackupDecryptionFailed
	}

	// 5. Decompress
	jsonData, err := compression.Decompress(compressed)
	if err != nil {
		return nil, fmt.Errorf("decompressing backup: %w", err)
	}

	// 6. Unmarshal
	var payload domain.BackupPayload
	if err := json.Unmarshal(jsonData, &payload); err != nil {
		return nil, fmt.Errorf("unmarshaling backup: %w", err)
	}

	return &payload, nil
}

// ---------------------------------------------------------------------------
// Data Restoration (ID remap → insert)
// ---------------------------------------------------------------------------

func (s *BackupService) insertRestoredData(
	ctx context.Context,
	userID primitive.ObjectID,
	payload *domain.BackupPayload,
) (*domain.Project, error) {
	// Build old → new ID mapping for all entities
	idMap := make(map[string]primitive.ObjectID)

	// 1. Create new Project
	newProjectID := primitive.NewObjectID()
	idMap[payload.Project.ID] = newProjectID

	now := time.Now().UTC()
	project := &domain.Project{
		ID:          newProjectID,
		Name:        payload.Project.Name,
		Description: payload.Project.Description,
		KeyEpoch:    payload.Project.KeyEpoch,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("creating project: %w", err)
	}

	// 2. Create member with keyrings from backup
	keyrings := make([]domain.ProjectMemberKeyring, len(payload.Member.Keyrings))
	for i, k := range payload.Member.Keyrings {
		keyrings[i] = domain.ProjectMemberKeyring{
			Epoch:                   k.Epoch,
			SecretPassphrase:        k.SecretPassphrase,
			SecretSigningPrivateKey: k.SecretSigningPrivateKey,
			SigningPublicKey:        k.SigningPublicKey,
		}
	}

	ownerMember := &domain.ProjectMember{
		ProjectID:           newProjectID,
		UserID:              userID,
		Role:                "owner",
		Permissions:         RolePresets["owner"],
		PublicKey:           payload.Member.PublicKey,
		EncryptedPrivateKey: payload.Member.EncryptedPrivateKey,
		Keyrings:            keyrings,
	}
	if err := s.memberRepo.Create(ctx, ownerMember); err != nil {
		return nil, fmt.Errorf("creating owner member: %w", err)
	}

	// 3. Pre-generate IDs for diagrams so parent references can be resolved
	for _, d := range payload.Diagrams {
		idMap[d.ID] = primitive.NewObjectID()
	}

	// Insert diagrams
	for _, d := range payload.Diagrams {
		diagram := &domain.Diagram{
			ID:                     idMap[d.ID],
			ProjectID:              newProjectID,
			DiagramName:            d.DiagramName,
			Description:            d.Description,
			EncryptedData:          d.EncryptedData,
			EncryptedDataSignature: d.EncryptedDataSignature,
		}
		if d.ParentDiagramID != nil {
			if newParent, ok := idMap[*d.ParentDiagramID]; ok {
				diagram.ParentDiagramID = &newParent
			}
		}
		if err := s.diagramRepo.Create(ctx, diagram); err != nil {
			return nil, fmt.Errorf("creating diagram %q: %w", d.DiagramName, err)
		}
	}

	// 4. Pre-generate IDs for nodes
	for _, n := range payload.Nodes {
		idMap[n.ID] = primitive.NewObjectID()
	}

	// Insert nodes
	for _, n := range payload.Nodes {
		node := &domain.Node{
			ID:                       idMap[n.ID],
			DiagramID:                idMap[n.DiagramID],
			EncryptedReadme:          n.EncryptedReadme,
			EncryptedReadmeSignature: n.EncryptedReadmeSignature,
			EncryptedDict:            n.EncryptedDict,
			EncryptedDictSignature:   n.EncryptedDictSignature,
		}
		if err := s.nodeRepo.Create(ctx, node); err != nil {
			return nil, fmt.Errorf("creating node: %w", err)
		}
	}

	// 5. Insert vaults
	for _, v := range payload.Vaults {
		vault := &domain.NodeVault{
			ProjectId:               newProjectID,
			NodeId:                  idMap[v.NodeID],
			Label:                   v.Label,
			Type:                    v.Type,
			EncryptedValue:          v.EncryptedValue,
			EncryptedValueSignature: v.EncryptedValueSignature,
		}
		if err := s.nodeVaultRepo.Create(ctx, vault); err != nil {
			return nil, fmt.Errorf("creating vault: %w", err)
		}
	}

	// 6. Pre-generate IDs for notes so parent references can be resolved
	for _, n := range payload.Notes {
		idMap[n.ID] = primitive.NewObjectID()
	}

	// Insert notes
	for _, n := range payload.Notes {
		note := &domain.Note{
			ID:                        idMap[n.ID],
			ProjectID:                 newProjectID,
			Type:                      n.Type,
			FileName:                  n.FileName,
			Icon:                      n.Icon,
			EncryptedContent:          n.EncryptedContent,
			EncryptedContentSignature: n.EncryptedContentSignature,
		}
		if n.ParentID != nil {
			if newParent, ok := idMap[*n.ParentID]; ok {
				note.ParentID = &newParent
			}
		}
		if err := s.noteRepo.Create(ctx, note); err != nil {
			return nil, fmt.Errorf("creating note %q: %w", n.FileName, err)
		}
	}

	return project, nil
}

// ---------------------------------------------------------------------------
// Domain → Backup Converters
// ---------------------------------------------------------------------------

func toProjectBackup(p *domain.Project) domain.ProjectBackup {
	return domain.ProjectBackup{
		ID:          p.ID.Hex(),
		Name:        p.Name,
		Description: p.Description,
		KeyEpoch:    p.KeyEpoch,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}

func toMemberBackup(m *domain.ProjectMember) domain.MemberBackup {
	keyrings := make([]domain.MemberKeyringBackup, len(m.Keyrings))
	for i, k := range m.Keyrings {
		keyrings[i] = domain.MemberKeyringBackup{
			Epoch:                   k.Epoch,
			SecretPassphrase:        k.SecretPassphrase,
			SecretSigningPrivateKey: k.SecretSigningPrivateKey,
			SigningPublicKey:        k.SigningPublicKey,
		}
	}
	return domain.MemberBackup{
		PublicKey:           m.PublicKey,
		EncryptedPrivateKey: m.EncryptedPrivateKey,
		Keyrings:            keyrings,
	}
}

func toDiagramBackups(diagrams []*domain.Diagram) []domain.DiagramBackup {
	result := make([]domain.DiagramBackup, len(diagrams))
	for i, d := range diagrams {
		result[i] = domain.DiagramBackup{
			ID:                     d.ID.Hex(),
			DiagramName:            d.DiagramName,
			Description:            d.Description,
			EncryptedData:          d.EncryptedData,
			EncryptedDataSignature: d.EncryptedDataSignature,
			CreatedAt:              d.CreatedAt.Format(time.RFC3339),
			UpdatedAt:              d.UpdatedAt.Format(time.RFC3339),
		}
		if d.ParentDiagramID != nil {
			hex := d.ParentDiagramID.Hex()
			result[i].ParentDiagramID = &hex
		}
	}
	return result
}

func toNodeBackups(nodes []*domain.Node) []domain.NodeBackup {
	result := make([]domain.NodeBackup, len(nodes))
	for i, n := range nodes {
		result[i] = domain.NodeBackup{
			ID:                       n.ID.Hex(),
			DiagramID:                n.DiagramID.Hex(),
			EncryptedReadme:          n.EncryptedReadme,
			EncryptedReadmeSignature: n.EncryptedReadmeSignature,
			EncryptedDict:            n.EncryptedDict,
			EncryptedDictSignature:   n.EncryptedDictSignature,
			CreatedAt:                n.CreatedAt.Format(time.RFC3339),
			UpdatedAt:                n.UpdatedAt.Format(time.RFC3339),
		}
	}
	return result
}

func toVaultBackups(vaults []*domain.NodeVault) []domain.VaultBackup {
	result := make([]domain.VaultBackup, len(vaults))
	for i, v := range vaults {
		result[i] = domain.VaultBackup{
			ID:                      v.ID.Hex(),
			NodeID:                  v.NodeId.Hex(),
			Label:                   v.Label,
			Type:                    v.Type,
			EncryptedValue:          v.EncryptedValue,
			EncryptedValueSignature: v.EncryptedValueSignature,
			CreatedAt:               v.CreatedAt.Format(time.RFC3339),
			UpdatedAt:               v.UpdatedAt.Format(time.RFC3339),
		}
	}
	return result
}

func toNoteBackups(notes []*domain.Note) []domain.NoteBackup {
	result := make([]domain.NoteBackup, len(notes))
	for i, n := range notes {
		result[i] = domain.NoteBackup{
			ID:                        n.ID.Hex(),
			Type:                      n.Type,
			FileName:                  n.FileName,
			Icon:                      n.Icon,
			EncryptedContent:          n.EncryptedContent,
			EncryptedContentSignature: n.EncryptedContentSignature,
			CreatedAt:                 n.CreatedAt.Format(time.RFC3339),
			UpdatedAt:                 n.UpdatedAt.Format(time.RFC3339),
		}
		if n.ParentID != nil {
			hex := n.ParentID.Hex()
			result[i].ParentID = &hex
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// toCryptoParams converts the service-level Argon2 params to the crypto
// package format, always using 32-byte (AES-256) key length.
func (s *BackupService) toCryptoParams() *crypto.Argon2Params {
	return &crypto.Argon2Params{
		Memory:      s.argon2Params.Memory,
		Iterations:  s.argon2Params.Iterations,
		Parallelism: s.argon2Params.Parallelism,
		KeyLength:   32,
	}
}

func sanitizeFilename(name string) string {
	result := make([]byte, 0, len(name))
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result = append(result, byte(c))
		} else if c == ' ' {
			result = append(result, '_')
		}
	}
	if len(result) == 0 {
		return "backup"
	}
	return string(result)
}
