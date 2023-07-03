package service

import (
	"context"
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/pyroscope-io/pyroscope/pkg/model"
	"github.com/pyroscope-io/pyroscope/pkg/model/appmetadata"
	"gorm.io/gorm"
)

type ApplicationMetadataService struct {
	db             *gorm.DB
	processingApps *sync.Map
}

func NewApplicationMetadataService(db *gorm.DB) *ApplicationMetadataService {
	svc := &ApplicationMetadataService{
		db:             db,
		processingApps: &sync.Map{},
	}
	go func() {
		svc.initLoadingApps()
	}()
	return svc
}

func (svc ApplicationMetadataService) initLoadingApps() {
	qry := url.Values{}
	qry.Set("updateTime", time.Now().Add(-time.Hour*12).Format("2006-01-02 15:04:05"))
	apps, err := svc.List(context.WithValue(context.Background(), "query", qry))
	if err != nil {
		return
	}
	for _, app := range apps {
		segmentKey := app.ToSegmentKey().Normalized()
		svc.processingApps.Store(segmentKey, struct{}{})
	}
}

func (svc ApplicationMetadataService) List(ctx context.Context) (apps []appmetadata.ApplicationMetadata, err error) {
	tx := svc.db.WithContext(ctx)
	query, ok := ctx.Value("query").(url.Values)
	if ok {
		if query.Get("projectID") != "" {
			tx = tx.Where("project_id = ?", query.Get("projectID"))
		}
		if query.Get("workspace") != "" {
			tx = tx.Where("workspace = ?", query.Get("workspace"))
		}
		if query.Get("orgID") != "" {
			tx = tx.Where("org_id = ?", query.Get("orgID"))
		}
		if query.Get("appID") != "" {
			tx = tx.Where("app_id = ?", query.Get("appID"))
		}
		if query.Get("podIP") != "" {
			tx = tx.Where("pod_ip = ?", query.Get("podIP"))
		}
		if query.Get("name") != "" {
			tx = tx.Where("name = ?", query.Get("name"))
		}
		if query.Get("updateTime") != "" {
			tx = tx.Where("updated_at >= ?", query.Get("updateTime"))
		}
	}
	result := tx.Find(&apps)
	return apps, result.Error
}

func (svc ApplicationMetadataService) Get(ctx context.Context, name string) (appmetadata.ApplicationMetadata, error) {
	app := appmetadata.ApplicationMetadata{}
	if err := model.ValidateAppName(name); err != nil {
		return app, err
	}

	tx := svc.db.WithContext(ctx)
	res := tx.Where("fq_name = ?", name).First(&app)

	switch {
	case errors.Is(res.Error, gorm.ErrRecordNotFound):
		return app, model.ErrApplicationNotFound
	default:
		return app, res.Error
	}
}

func (svc ApplicationMetadataService) CreateOrUpdate(ctx context.Context, application appmetadata.ApplicationMetadata) error {
	segmentKey := application.ToSegmentKey().Normalized()
	if _, ok := svc.processingApps.Load(segmentKey); ok {
		return nil
	}
	if err := model.ValidateAppName(application.FQName); err != nil {
		return err
	}

	tx := svc.db.WithContext(ctx)

	// Only update the field if it's populated
	if err := tx.Where(appmetadata.ApplicationMetadata{
		FQName:      application.FQName,
		ProjectID:   application.ProjectID,
		ProjectName: application.ProjectName,
		OrgID:       application.OrgID,
		OrgName:     application.OrgName,
		Workspace:   application.Workspace,
		AppID:       application.AppID,
		SpyName:     application.SpyName,
		ServiceName: application.ServiceName,
		PodIP:       application.PodIP,
	}).Assign(application).FirstOrCreate(&appmetadata.ApplicationMetadata{}).Error; err != nil {
		return err
	}
	svc.processingApps.Store(segmentKey, struct{}{})
	return nil
}

func (svc ApplicationMetadataService) Delete(ctx context.Context, name string) error {
	if err := model.ValidateAppName(name); err != nil {
		return err
	}

	tx := svc.db.WithContext(ctx)
	return tx.Where("fq_name = ?", name).Delete(appmetadata.ApplicationMetadata{}).Error
}
