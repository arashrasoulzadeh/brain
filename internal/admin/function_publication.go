package admin

import (
	"encoding/json"
	"errors"
	"maps"
	"os"
	"strings"

	"brain/internal/execution"
)

var ErrFunctionVersionConflict = errors.New("published function version already pinned")

const maxFunctionPublications = 128

// FunctionPublication is an immutable source snapshot. Draft metadata and
// visibility are independent from the selected exact execution version.
type FunctionPublication struct {
	Version int    `json:"version"`
	Source  string `json:"source"`
	Method  string `json:"method"`
	Digest  string `json:"digest"`
}

func pinFunctionPublication(draft FunctionDraft) (FunctionDraft, error) {
	method, err := functionMethodFromSource(draft.Source)
	if err != nil {
		return draft, err
	}
	if draft.Version < 1 || len(draft.Source) > maxFunctionSourceBytes || !functionExportPattern.MatchString(draft.Source) || functionUnsafePattern.MatchString(draft.Source) {
		return draft, ErrFunctionSourceInvalid
	}
	pin := FunctionPublication{Version: draft.Version, Source: draft.Source, Method: method, Digest: execution.ArtifactDigest([]byte(draft.Source))}
	if existing, ok := draft.PublishedVersions[draft.Version]; ok {
		if existing != pin {
			return draft, ErrFunctionVersionConflict
		}
	} else {
		if len(draft.PublishedVersions) >= maxFunctionPublications {
			return draft, ErrFunctionVersionConflict
		}
		// Copy before mutation so a failed atomic file write cannot alter the
		// previously admitted publication map through a shared reference.
		pins := maps.Clone(draft.PublishedVersions)
		if pins == nil {
			pins = make(map[int]FunctionPublication)
		}
		pins[draft.Version] = pin
		draft.PublishedVersions = pins
	}
	draft.ActivePublication = draft.Version
	return draft, nil
}

// pinNextFunctionPublication preserves existing pins and allocates a new
// version only when an explicit publication would change an existing pin.
// The caller holds functionDrafts.Lock across selection and persistence.
func pinNextFunctionPublication(draft FunctionDraft) (FunctionDraft, error) {
	pinned, err := pinFunctionPublication(draft)
	if !errors.Is(err, ErrFunctionVersionConflict) {
		return pinned, err
	}
	if _, occupied := draft.PublishedVersions[draft.Version]; !occupied || len(draft.PublishedVersions) >= maxFunctionPublications {
		return draft, err
	}
	highest := draft.Version
	for version := range draft.PublishedVersions {
		if version > highest {
			highest = version
		}
	}
	// Avoid integer overflow and version numbers that cannot round-trip
	// through the JavaScript authoring client.
	if highest == int(^uint(0)>>1) || int64(highest) >= 9007199254740991 {
		return draft, ErrFunctionVersionConflict
	}
	draft.Version = highest + 1
	return pinFunctionPublication(draft)
}

// Legacy records did not preserve the published bytes. Adopt the first known
// source before editing; do not pretend historical publications can be restored.
func adoptLegacyPublication(draft FunctionDraft) FunctionDraft {
	if draft.ActivePublication != 0 || len(draft.PublishedVersions) != 0 || (!draft.Public && draft.PublishedAt == nil) {
		return draft
	}
	if draft.Version == 0 {
		draft.Version = 1
	}
	pinned, err := pinFunctionPublication(draft)
	if err != nil {
		return draft
	}
	return pinned
}

// Public execution reads durable state rather than a stale in-memory draft
// cache. A missing, corrupt, revoked or ambiguous publication fails closed.
func loadPublicFunctionPublication(projectID, name string) (FunctionDraft, FunctionPublication, error) {
	functionDrafts.Lock()
	defer functionDrafts.Unlock()
	raw, err := os.ReadFile(functionDraftsPath)
	if err != nil {
		return FunctionDraft{}, FunctionPublication{}, ErrFunctionDraftInvalid
	}
	var drafts []FunctionDraft
	if err := json.Unmarshal(raw, &drafts); err != nil {
		return FunctionDraft{}, FunctionPublication{}, ErrFunctionDraftInvalid
	}
	var found FunctionDraft
	for _, draft := range drafts {
		if draft.ProjectID == projectID && draft.Name == strings.TrimSpace(name) {
			if found.Name != "" {
				return FunctionDraft{}, FunctionPublication{}, ErrFunctionDraftInvalid
			}
			found = adoptLegacyPublication(draft)
		}
	}
	if found.Name == "" || !found.Public || found.ActivePublication < 1 {
		return FunctionDraft{}, FunctionPublication{}, ErrFunctionDraftInvalid
	}
	publication, ok := found.PublishedVersions[found.ActivePublication]
	if !ok || publication.Version != found.ActivePublication || publication.Digest != execution.ArtifactDigest([]byte(publication.Source)) {
		return FunctionDraft{}, FunctionPublication{}, ErrFunctionDraftInvalid
	}
	method, err := functionMethodFromSource(publication.Source)
	if err != nil || method != publication.Method {
		return FunctionDraft{}, FunctionPublication{}, ErrFunctionDraftInvalid
	}
	return found, publication, nil
}
