package wmf

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/stretchr/testify/mock"
)

var _ API = (*MockWmf)(nil)

type MockWmf struct {
	mock.Mock
}

func GetMock() API {
	return &MockWmf{}
}
func (m *MockWmf) GetWikidataRevertRiskScore(ctx context.Context, rev int) (*WikidataRevertRiskScore, error) {
	args := m.Called(ctx, rev)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*WikidataRevertRiskScore), args.Error(1)
}

func (m *MockWmf) GetWikidataEntityCreationDate(ctx context.Context, entityTitle string) (*time.Time, error) {
	args := m.Called(ctx, entityTitle)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*time.Time), args.Error(1)
}

func (m *MockWmf) GetWikibaseEntity(ctx context.Context, dtb string, entityID string, options ...func(*url.Values)) (*WikibaseEntityResponse, error) {
	args := m.Called(ctx, dtb, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*WikibaseEntityResponse), args.Error(1)
}

func (m *MockWmf) GetAllPages(ctx context.Context, dtb string, cbk func([]*Page), ops ...func(*url.Values)) error {
	args := m.Called(ctx, dtb)

	if args.Get(0) != nil {
		pages := args.Get(0).([]*Page)
		cbk(pages)
	}

	return args.Error(1)
}

func (m *MockWmf) GetPages(ctx context.Context, dtb string, tls []string, ops ...func(*url.Values)) (map[string]*Page, error) {
	args := m.Called(ctx, dtb, tls)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*Page), args.Error(1)
}

func (m *MockWmf) GetPage(ctx context.Context, dtb string, ttl string, ops ...func(*url.Values)) (*Page, error) {
	args := m.Called(ctx, dtb, ttl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *MockWmf) GetPageHTML(ctx context.Context, dtb string, ttl string, ops ...func(*url.Values)) (string, error) {
	args := m.Called(ctx, dtb, ttl)
	return args.String(0), args.Error(1)
}

func (m *MockWmf) GetPagesHTML(ctx context.Context, dtb string, tls []string, mxc int, ops ...func(*url.Values)) map[string]*PageHTML {
	args := m.Called(ctx, dtb, tls, mxc)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[string]*PageHTML)
}

func (m *MockWmf) GetRevisionHTML(ctx context.Context, dtb string, rid string, ops ...func(*url.Values)) (string, error) {
	args := m.Called(ctx, dtb, rid)
	return args.String(0), args.Error(1)
}

func (m *MockWmf) GetRevisionsHTML(ctx context.Context, dtb string, rvs []string, mxc int, ops ...func(*url.Values)) map[string]*PageHTML {
	args := m.Called(ctx, dtb, rvs, mxc)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[string]*PageHTML)
}

func (m *MockWmf) GetPageByRevision(ctx context.Context, dtb string, revid int, ops ...func(*url.Values)) (*Page, error) {
	args := m.Called(ctx, dtb, revid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *MockWmf) GetAllRevisions(ctx context.Context, dtb string, pageid int, ops ...func(*url.Values)) ([]*Revision, error) {
	args := m.Called(ctx, dtb, pageid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Revision), args.Error(1)
}

func (m *MockWmf) GetLanguages(ctx context.Context, dtb string, ops ...func(*url.Values)) ([]*Language, error) {
	args := m.Called(ctx, dtb)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Language), args.Error(1)
}

func (m *MockWmf) GetLanguage(ctx context.Context, dtb string) (*Language, error) {
	args := m.Called(ctx, dtb)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Language), args.Error(1)
}

func (m *MockWmf) GetProject(ctx context.Context, dtb string) (*Project, error) {
	args := m.Called(ctx, dtb)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Project), args.Error(1)
}

func (m *MockWmf) GetNamespaces(ctx context.Context, dtb string, ops ...func(*url.Values)) ([]*Namespace, error) {
	args := m.Called(ctx, dtb)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Namespace), args.Error(1)
}

func (m *MockWmf) GetRandomPages(ctx context.Context, dtb string, ops ...func(*url.Values)) ([]*Page, error) {
	args := m.Called(ctx, dtb)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Page), args.Error(1)
}

func (m *MockWmf) GetUsers(ctx context.Context, dtb string, ids []int, ops ...func(*url.Values)) (map[int]*User, error) {
	args := m.Called(ctx, dtb, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int]*User), args.Error(1)
}

func (m *MockWmf) GetUser(ctx context.Context, dtb string, id int, ops ...func(*url.Values)) (*User, error) {
	args := m.Called(ctx, dtb, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockWmf) GetScore(ctx context.Context, rev int, lng string, prj string, mdl string) (*Score, error) {
	args := m.Called(ctx, rev, lng, prj, mdl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Score), args.Error(1)
}

func (m *MockWmf) GetReferenceNeedScore(ctx context.Context, rev int, lng, prj string) (*ReferenceNeedScore, error) {
	args := m.Called(ctx, rev, lng, prj)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReferenceNeedScore), args.Error(1)
}

func (m *MockWmf) GetReferenceRiskScore(ctx context.Context, rev int, lng, prj string) (*ReferenceRiskScore, error) {
	args := m.Called(ctx, rev, lng, prj)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReferenceRiskScore), args.Error(1)
}

func (m *MockWmf) GetPageSummary(ctx context.Context, dtb string, ttl string, ops ...func(*url.Values)) (*PageSummary, error) {
	args := m.Called(ctx, dtb, ttl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PageSummary), args.Error(1)
}

func (m *MockWmf) GetContributors(ctx context.Context, dtb string, pageid int, maxRegistered *int, ops ...func(*url.Values)) ([]*Page, bool, error) {
	var limit any = nil
	if maxRegistered != nil {
		limit = *maxRegistered
	}
	args := m.Called(ctx, dtb, pageid, limit)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).([]*Page), args.Bool(1), args.Error(2)
}

func (m *MockWmf) GetContributorsCount(ctx context.Context, dtb string, pageid int, maxRegistered *int, ops ...func(*url.Values)) (*ContributorsCount, error) {
	args := m.Called(ctx, dtb, pageid, maxRegistered)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ContributorsCount), args.Error(1)
}

func (m *MockWmf) DownloadFile(ctx context.Context, url string, options ...func(*http.Request)) ([]byte, error) {
	args := m.Called(ctx, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockWmf) HeadFile(ctx context.Context, url string, options ...func(*http.Request)) ([]byte, error) {
	args := m.Called(ctx, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
