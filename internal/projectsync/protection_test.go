package projectsync

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func (c *fakeSourceClient) ProtectedTags(context.Context, Source, Project) ([]string, error) {
	return nil, nil
}
func (c *fakeSourceClient) ProtectBranch(context.Context, Source, Project, string) error { return nil }
func (c *fakeSourceClient) ProtectTag(context.Context, Source, Project, string) error    { return nil }

func TestProtectionChecksAndOperations(t *testing.T) {
	for _, tags := range []bool{false, true} {
		t.Run(fmt.Sprint("tags=", tags), func(t *testing.T) {
			required := requiredProtectedBranches
			endpoint := "protected_branches"
			checkID, operationID := checkProtectedBranches, operationProtectBranches
			if tags {
				required, endpoint = requiredProtectedTags, "protected_tags"
				checkID, operationID = checkProtectedTags, operationProtectTags
			}
			names := []string{required[0], "unrelated"}
			posts := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v4/projects/42/"+endpoint {
					t.Errorf("unexpected path %s", r.URL.Path)
					http.NotFound(w, r)
					return
				}
				if r.Header.Get("PRIVATE-TOKEN") != "token" {
					t.Error("missing token")
				}
				switch r.Method {
				case http.MethodGet:
					raw := make([]map[string]string, 0, len(names))
					for _, name := range names {
						raw = append(raw, map[string]string{"name": name})
					}
					_ = json.NewEncoder(w).Encode(raw)
				case http.MethodPost:
					_ = r.ParseForm()
					name := r.Form.Get("name")
					for _, existing := range names {
						if name == existing {
							t.Error("recreated existing protection")
						}
					}
					if tags {
						if r.Form.Get("create_access_level") != "40" {
							t.Error("incorrect tag access")
						}
					} else {
						if r.Form.Get("push_access_level") != "40" || r.Form.Get("merge_access_level") != "30" || r.Form.Get("allow_force_push") != "false" {
							t.Error("incorrect branch access")
						}
					}
					names = append(names, name)
					posts++
					w.WriteHeader(http.StatusCreated)
				default:
					t.Errorf("unexpected method %s", r.Method)
				}
			}))
			defer server.Close()
			client := GitLabClient{HTTPClient: server.Client()}
			source, project := Source{Type: SourceTypeGitLab, URL: server.URL, PATToken: "token"}, Project{ProviderID: "42"}
			checks := CheckService{SourceClient: client}
			state, err := checks.runProtectionCheck(context.Background(), source, project, checkID)
			if err != nil || state != CheckStateFail {
				t.Fatalf("before: %s %v", state, err)
			}
			operations := OperationService{SourceClient: client}
			for i := 0; i < 2; i++ {
				if err := operations.Run(context.Background(), source, project, operationID); err != nil {
					t.Fatal(err)
				}
			}
			if posts != len(required)-1 {
				t.Fatalf("posts = %d", posts)
			}
			state, err = checks.runProtectionCheck(context.Background(), source, project, checkID)
			if err != nil || state != CheckStatePass {
				t.Fatalf("after: %s %v", state, err)
			}
		})
	}
}

func TestProtectionChecksNonGitLab(t *testing.T) {
	for _, id := range []string{checkProtectedBranches, checkProtectedTags} {
		state, err := (CheckService{}).runProtectionCheck(context.Background(), Source{Type: "github"}, Project{}, id)
		if err != nil || state != CheckStateNotApplicable {
			t.Fatalf("%s: %s %v", id, state, err)
		}
	}
}

func TestProtectionErrorsAreReported(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusForbidden) }))
	defer server.Close()
	client := GitLabClient{HTTPClient: server.Client()}
	source, project := Source{Type: SourceTypeGitLab, URL: server.URL}, Project{ProviderID: "42"}
	for _, id := range []string{checkProtectedBranches, checkProtectedTags} {
		if _, err := (CheckService{SourceClient: client}).runProtectionCheck(context.Background(), source, project, id); err == nil {
			t.Fatal("expected read error")
		}
	}
	for _, id := range []string{operationProtectBranches, operationProtectTags} {
		if err := (OperationService{SourceClient: client}).Run(context.Background(), source, project, id); err == nil {
			t.Fatal("expected operation error")
		}
	}
}

func TestProtectionListsReadAllPages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("per_page") != "100" {
			t.Error("missing page size")
		}
		var raw []map[string]string
		switch r.URL.Query().Get("page") {
		case "1":
			for i := 0; i < 100; i++ {
				raw = append(raw, map[string]string{"name": fmt.Sprintf("other-%d", i)})
			}
		case "2":
			raw = append(raw, map[string]string{"name": "last"})
		default:
			t.Errorf("unexpected page %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(raw)
	}))
	defer server.Close()
	client := GitLabClient{HTTPClient: server.Client()}
	source, project := Source{URL: server.URL}, Project{ProviderID: "42"}
	tags, err := client.ProtectedTags(context.Background(), source, project)
	if err != nil || len(tags) != 101 || tags[100] != "last" {
		t.Fatalf("tags: %v %v", tags, err)
	}
	branches, err := client.ProtectedBranches(context.Background(), source, project)
	if err != nil || len(branches) != 101 || branches[100].Name != "last" {
		t.Fatalf("branches: %v %v", branches, err)
	}
}
