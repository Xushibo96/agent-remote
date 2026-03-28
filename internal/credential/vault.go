package credential

import (
	"context"
	"fmt"
	"time"

	toolerrors "agent-remote/internal/errors"
	"agent-remote/internal/model"
	"agent-remote/internal/secret"
	"agent-remote/internal/store"
)

type ConfigStore interface {
	SaveTarget(ctx context.Context, target model.TargetConfig) error
	GetTarget(ctx context.Context, targetID string) (model.TargetConfig, error)
	ListTargets(ctx context.Context) ([]model.TargetConfig, error)
}

type Vault struct {
	store      ConfigStore
	keyManager *secret.KeyManager
}

func NewVault(store ConfigStore, keyManager *secret.KeyManager) *Vault {
	return &Vault{store: store, keyManager: keyManager}
}

func (v *Vault) SaveTarget(ctx context.Context, req model.AddTargetRequest) (model.TargetConfig, error) {
	normalized, err := model.NormalizeAddTargetRequest(req)
	if err != nil {
		return model.TargetConfig{}, &toolerrors.ToolError{
			Code:      "invalid_target",
			Category:  "validation",
			Stage:     "normalize_target",
			Message:   err.Error(),
			Retryable: false,
		}
	}

	target := model.TargetConfig{
		ID:               normalized.ID,
		Name:             normalized.Name,
		Host:             normalized.Host,
		Port:             normalized.Port,
		User:             normalized.User,
		AuthMode:         normalized.AuthMode,
		PrivateKeyPath:   normalized.PrivateKeyPath,
		KnownHostsPolicy: normalized.KnownHostsPolicy,
		DefaultBaseDir:   normalized.DefaultBaseDir,
		Tags:             normalized.Tags,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	if normalized.AuthMode == model.AuthModePassword {
		key, keyID, err := v.keyManager.LoadOrCreateMasterKey()
		if err != nil {
			return model.TargetConfig{}, &toolerrors.ToolError{
				Code:        "keyring_unavailable",
				Category:    "auth",
				Stage:       "load_master_key",
				Message:     "failed to load or create master key",
				Retryable:   false,
				Remediation: "verify local keyring availability",
			}
		}
		envelope, err := secret.Encrypt(key, normalized.Password, keyID)
		if err != nil {
			return model.TargetConfig{}, &toolerrors.ToolError{
				Code:      "encrypt_secret_failed",
				Category:  "auth",
				Stage:     "encrypt_secret",
				Message:   "failed to encrypt target password",
				Retryable: false,
			}
		}
		target.PasswordEnvelope = envelope
	}

	if err := v.store.SaveTarget(ctx, target); err != nil {
		return model.TargetConfig{}, err
	}
	return target, nil
}

func (v *Vault) LoadTarget(ctx context.Context, targetID string) (model.ResolvedTarget, error) {
	target, err := v.store.GetTarget(ctx, targetID)
	if err != nil {
		if err == store.ErrTargetNotFound {
			return model.ResolvedTarget{}, &toolerrors.ToolError{
				Code:      "target_not_found",
				Category:  "validation",
				Stage:     "load_target",
				Message:   fmt.Sprintf("target %q was not found", targetID),
				Retryable: false,
			}
		}
		return model.ResolvedTarget{}, err
	}

	resolved := model.ResolvedTarget{TargetConfig: target}
	if target.AuthMode != model.AuthModePassword {
		return resolved, nil
	}

	key, _, err := v.keyManager.LoadOrCreateMasterKey()
	if err != nil {
		return model.ResolvedTarget{}, &toolerrors.ToolError{
			Code:        "keyring_unavailable",
			Category:    "auth",
			Stage:       "load_master_key",
			Message:     "failed to load master key",
			Retryable:   false,
			Remediation: "verify local keyring availability",
		}
	}
	password, err := secret.Decrypt(key, target.PasswordEnvelope)
	if err != nil {
		return model.ResolvedTarget{}, &toolerrors.ToolError{
			Code:        "credential_decrypt_failed",
			Category:    "auth",
			Stage:       "decrypt_secret",
			Message:     "failed to decrypt stored password",
			Retryable:   false,
			Remediation: "update the target password and save again",
		}
	}
	resolved.Password = password
	return resolved, nil
}

func (v *Vault) ListTargets(ctx context.Context) ([]model.TargetConfig, error) {
	return v.store.ListTargets(ctx)
}

func (v *Vault) RotateMasterKey(context.Context) error {
	return fmt.Errorf("not implemented")
}
