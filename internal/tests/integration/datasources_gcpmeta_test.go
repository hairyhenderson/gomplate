package integration

import (
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// setupGCPMetaTest creates an HTTP server that serves mock GCP metadata.
// The Google metadata client will talk to this server instead of the real
// metadata service.
func setupGCPMetaTest(t *testing.T) *httptest.Server {
	t.Helper()

	metafsys := createGCPMockFS()
	mux := http.NewServeMux()

	// Handle the computeMetadata requests
	mux.HandleFunc("/computeMetadata/v1/", handleGCPMetadataRequest(metafsys))

	// Handle the special token URL that the client uses to authenticate
	mux.HandleFunc("/computeMetadata/v1/instance/service-accounts/default/token",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Metadata-Flavor") != "Google" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			_, _ = w.Write([]byte(`{"access_token":"fake-token","expires_in":3600,"token_type":"Bearer"}`))
		})

	srv := httptest.NewServer(checkGCPMetadataFlavor(mux))
	t.Cleanup(srv.Close)

	return srv
}

// checkGCPMetadataFlavor adds middleware to check for the Google Metadata
// flavor header and handle redirects for missing trailing slashes
func checkGCPMetadataFlavor(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for required Metadata-Flavor header
		if r.Header.Get("Metadata-Flavor") != "Google" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		wrec := httptest.NewRecorder()
		handler.ServeHTTP(wrec, r)

		// try again on 301s - likely just a trailing `/` missing
		if wrec.Code == http.StatusMovedPermanently {
			if !strings.HasSuffix(r.URL.Path, "/") {
				r.URL.Path += "/"
			}
			handler.ServeHTTP(w, r)
			return
		}

		maps.Copy(w.Header(), wrec.Header())
		w.WriteHeader(wrec.Code)
		_, _ = w.Write(wrec.Body.Bytes())
	})
}

// handleGCPMetadataRequest returns a handler for metadata requests that serves
// content from the mock filesystem
func handleGCPMetadataRequest(metafsys fstest.MapFS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Strip the prefix to get the actual path
		path := strings.TrimPrefix(r.URL.Path, "/computeMetadata/v1/")

		// Root path special handling
		if path == "" {
			_, _ = w.Write([]byte("instance/\nproject/\n"))
			return
		}

		// Special case for ReadDir tests
		if path == "instance/" {
			dirContent := []string{
				"attributes/", "cpu-platform", "disks/", "hostname", "id", "image",
				"machine-type", "network-interfaces/", "service-accounts/", "zone",
			}
			_, _ = w.Write([]byte(strings.Join(dirContent, "\n") + "\n"))
			return
		}

		// Special cases for other directory paths
		if path == "project/" {
			_, _ = w.Write([]byte("attributes/\nnumeric-project-id\nproject-id\n"))
			return
		}

		if path == "project/attributes/" {
			_, _ = w.Write([]byte("my-custom-attribute\nssh-keys\n"))
			return
		}

		if path == "instance/network-interfaces/" {
			_, _ = w.Write([]byte("0/\n"))
			return
		}

		if path == "instance/service-accounts/" {
			_, _ = w.Write([]byte("default/\n"))
			return
		}

		// Check if it's a directory request (ends with /)
		if strings.HasSuffix(path, "/") {
			entries := getGCPDirectoryEntries(metafsys, path)
			if len(entries) == 0 {
				// Directory not found
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "GCE metadata %q not defined", path)
				return
			}
			_, _ = w.Write([]byte(strings.Join(entries, "\n")))
			return
		}

		// Check if the file exists
		if file, ok := metafsys[path]; ok {
			_, _ = w.Write(file.Data)
			return
		}

		// File not found
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "GCE metadata %q not defined", path)
	}
}

// getGCPDirectoryEntries extracts directory entries for a given path prefix
func getGCPDirectoryEntries(fs fstest.MapFS, prefix string) []string {
	entries := map[string]struct{}{}

	for name := range fs {
		if !strings.HasPrefix(name, prefix) {
			continue
		}

		relPath := strings.TrimPrefix(name, prefix)
		if relPath == "" {
			continue
		}

		entryName := relPath
		if idx := strings.Index(relPath, "/"); idx >= 0 {
			entryName = relPath[:idx+1]
		}

		entries[entryName] = struct{}{}
	}

	result := make([]string, 0, len(entries))
	for entry := range entries {
		result = append(result, entry)
	}

	return result
}

// createGCPMockFS creates the fake metadata filesystem for testing
func createGCPMockFS() fstest.MapFS {
	return fstest.MapFS{
		// Instance metadata values
		"instance/attributes/enable-oslogin": &fstest.MapFile{Data: []byte("FALSE")},
		"instance/attributes/shell":          &fstest.MapFile{Data: []byte("bash")},
		"instance/cpu-platform":              &fstest.MapFile{Data: []byte("Intel Haswell")},
		"instance/hostname":                  &fstest.MapFile{Data: []byte("instance-1.c.test-project.internal")},
		"instance/id":                        &fstest.MapFile{Data: []byte("1234567890123456789")},
		"instance/image":                     &fstest.MapFile{Data: []byte("projects/debian-cloud/global/images/debian-12")},
		"instance/machine-type":              &fstest.MapFile{Data: []byte("projects/123456789012/machineTypes/e2-medium")},
		"instance/zone":                      &fstest.MapFile{Data: []byte("projects/123456789012/zones/us-central1-a")},

		// Network interfaces
		"instance/network-interfaces/0/ip":          &fstest.MapFile{Data: []byte("10.128.0.2")},
		"instance/network-interfaces/0/mac":         &fstest.MapFile{Data: []byte("42:01:0a:80:00:02")},
		"instance/network-interfaces/0/network":     &fstest.MapFile{Data: []byte("projects/123456789012/networks/default")},
		"instance/network-interfaces/0/external-ip": &fstest.MapFile{Data: []byte("34.123.45.67")},

		// Service accounts
		"instance/service-accounts/default/aliases": &fstest.MapFile{Data: []byte("default")},
		"instance/service-accounts/default/email":   &fstest.MapFile{Data: []byte("123456789012-compute@developer.gserviceaccount.com")},
		"instance/service-accounts/default/scopes":  &fstest.MapFile{Data: []byte("https://www.googleapis.com/auth/cloud-platform")},

		// Project metadata values
		"project/numeric-project-id":             &fstest.MapFile{Data: []byte("123456789012")},
		"project/project-id":                     &fstest.MapFile{Data: []byte("test-project-id")},
		"project/attributes/my-custom-attribute": &fstest.MapFile{Data: []byte("custom-value")},
		"project/attributes/ssh-keys":            &fstest.MapFile{Data: []byte("user:ssh-rsa AAAAB3... test")},
	}
}

func TestDatasources_GCPMeta_Instance(t *testing.T) {
	srv := setupGCPMetaTest(t)

	// Trim http:// prefix - the GCE_METADATA_HOST expects host:port only
	metadataHost := strings.TrimPrefix(srv.URL, "http://")

	t.Run("read instance id", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "meta=gcp+meta:///",
			"-i", `{{ include "meta" "instance/id" }}`).
			withEnv("GCE_METADATA_HOST", metadataHost).run()
		assertSuccess(t, o, e, err, "1234567890123456789")
	})

	t.Run("read instance hostname", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "meta=gcp+meta:///",
			"-i", `{{ include "meta" "instance/hostname" }}`).
			withEnv("GCE_METADATA_HOST", metadataHost).run()
		assertSuccess(t, o, e, err, "instance-1.c.test-project.internal")
	})

	t.Run("read instance zone", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "meta=gcp+meta:///",
			"-i", `{{ include "meta" "instance/zone" }}`).
			withEnv("GCE_METADATA_HOST", metadataHost).run()
		assertSuccess(t, o, e, err, "projects/123456789012/zones/us-central1-a")
	})

	t.Run("read instance cpu-platform", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "meta=gcp+meta:///",
			"-i", `{{ include "meta" "instance/cpu-platform" }}`).
			withEnv("GCE_METADATA_HOST", metadataHost).run()
		assertSuccess(t, o, e, err, "Intel Haswell")
	})

	t.Run("read network interface ip", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "meta=gcp+meta:///",
			"-i", `{{ include "meta" "instance/network-interfaces/0/ip" }}`).
			withEnv("GCE_METADATA_HOST", metadataHost).run()
		assertSuccess(t, o, e, err, "10.128.0.2")
	})

	t.Run("read instance attribute", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "meta=gcp+meta:///",
			"-i", `{{ include "meta" "instance/attributes/shell" }}`).
			withEnv("GCE_METADATA_HOST", metadataHost).run()
		assertSuccess(t, o, e, err, "bash")
	})
}

func TestDatasources_GCPMeta_Project(t *testing.T) {
	srv := setupGCPMetaTest(t)
	metadataHost := strings.TrimPrefix(srv.URL, "http://")

	t.Run("read project-id", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "meta=gcp+meta:///",
			"-i", `{{ include "meta" "project/project-id" }}`).
			withEnv("GCE_METADATA_HOST", metadataHost).run()
		assertSuccess(t, o, e, err, "test-project-id")
	})

	t.Run("read numeric-project-id", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "meta=gcp+meta:///",
			"-i", `{{ include "meta" "project/numeric-project-id" }}`).
			withEnv("GCE_METADATA_HOST", metadataHost).run()
		assertSuccess(t, o, e, err, "123456789012")
	})

	t.Run("read custom project attribute", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "meta=gcp+meta:///",
			"-i", `{{ include "meta" "project/attributes/my-custom-attribute" }}`).
			withEnv("GCE_METADATA_HOST", metadataHost).run()
		assertSuccess(t, o, e, err, "custom-value")
	})
}

func TestDatasources_GCPMeta_Directory(t *testing.T) {
	srv := setupGCPMetaTest(t)
	metadataHost := strings.TrimPrefix(srv.URL, "http://")

	t.Run("list root directory", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "meta=gcp+meta:///",
			"-i", `{{ ds "meta" }}`).
			withEnv("GCE_METADATA_HOST", metadataHost).run()
		assertSuccess(t, o, e, err, "[instance project]")
	})

	t.Run("list project directory", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "meta=gcp+meta:///project/",
			"-i", `{{ coll.Has (ds "meta") "project-id" }}`).
			withEnv("GCE_METADATA_HOST", metadataHost).run()
		assertSuccess(t, o, e, err, "true")
	})
}

func TestDatasources_GCPMeta_Context(t *testing.T) {
	srv := setupGCPMetaTest(t)
	metadataHost := strings.TrimPrefix(srv.URL, "http://")

	t.Run("use context flag", func(t *testing.T) {
		o, e, err := cmd(t,
			"-c", "inst=gcp+meta:///",
			"-i", `{{ include "inst" "instance/id" }}`).
			withEnv("GCE_METADATA_HOST", metadataHost).run()
		assertSuccess(t, o, e, err, "1234567890123456789")
	})
}
