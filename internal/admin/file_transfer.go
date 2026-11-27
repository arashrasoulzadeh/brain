package admin

import (
	"context"
	"io"

	"brain/internal/operator"
	"brain/internal/plugin"
)

func (s *LifecycleService) UploadFile(actor operator.Actor, ctx context.Context, projectID, version, key, filename, contentType string, source io.Reader, size int64) (plugin.FileMetadata, error) {
	var metadata plugin.FileMetadata
	err := s.execute(actor, "plugin.admin.invoke", func() error {
		invocation := PluginInvocation{ProjectID: projectID, PluginID: "files", Version: version, Operation: "files.upload"}
		if _, err := s.adminPageManifest(actor, "operations", invocation); err != nil {
			return err
		}
		var err error
		metadata, err = s.registry.UploadFileAsOperator(ctx, projectID, version, key, actor.ID, filename, contentType, source, size, func() error {
			return s.operators.AuthorizeAction(actor, "plugin.admin.invoke")
		})
		return err
	})
	return metadata, err
}

func (s *LifecycleService) DownloadFile(actor operator.Actor, ctx context.Context, projectID, version, key string, begin func(plugin.FileMetadata) io.Writer) error {
	return s.execute(actor, "plugin.admin.query", func() error {
		invocation := PluginInvocation{ProjectID: projectID, PluginID: "files", Version: version, Operation: "files.download"}
		manifest, err := s.adminPageManifest(actor, "operations", invocation)
		if err != nil {
			return err
		}
		if !manifest.AdminReadOnly("operations", "files.download") {
			return plugin.ErrGrantDenied
		}
		return s.registry.DownloadFileAsOperator(ctx, projectID, version, key, actor.ID, begin)
	})
}
