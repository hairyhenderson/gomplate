package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hairyhenderson/gomplate/v4/vault"
	"github.com/rs/zerolog"
)

// get the vault mount point and kv version of a path using /sys/internal/ui/mounts API call
func getPathMetadataInternal(log *zerolog.Logger, vc *vault.Vault, path string) (mountPoint string, version string, err error) {
	mountPoint = ""
	version = ""
	jsonData, err := vc.Read(fmt.Sprintf("/sys/internal/ui/mounts%s", path))
	if err == nil {
		var decodedData map[string]any
		json.Unmarshal(jsonData, &decodedData)
		mountPoint = decodedData["path"].(string)
		version = "1"
		if decodedData["options"] != nil {
			options := decodedData["options"].(map[string]any)
			version = options["version"].(string)
		}
		mountPoint = fmt.Sprintf("/%s", mountPoint)
	}
	log.Debug().Err(err).Msgf("readVault: getPathMetadataInternal:\n mountPoint: %s\n version: %s", mountPoint, version)
	return mountPoint, version, err
}

// get the vault mount point of a path by iterating all the mount points and
// choosing the one which is the longest prefix of the path
func getMountPoint(mpList map[string]any, path string) (mountPoint string, err error) {
	x := 0
	y := 0
	for mp := range mpList {
		mp = fmt.Sprintf("/%s", mp)
		if strings.HasPrefix(path, mp) {
			y = len([]rune(mp))
			if y > x {
				mountPoint = mp
				x = y
			}
		}
	}
	if len(mountPoint) == 0 {
		err = errors.New("mount point not found")
	}
	return mountPoint, err
}

// get the vault mount point of a path using getMountPoint and
// version by testing MOUNTPOINT/metadata/RELATIVE-PATH (only kv2 has metadata)
func getPathMetadataPrefix(log *zerolog.Logger, vc *vault.Vault, path string) (mountPoint string, version string, err error) {
	mountPoint = ""
	version = ""
	jsonData, err := vc.Read("/sys/mounts")
	if err == nil {
		var decodedData map[string]any
		json.Unmarshal(jsonData, &decodedData)
		mountPoint, err = getMountPoint(decodedData, path)
		if err == nil {
			_, err = vc.Read(fmt.Sprintf("%smetadata/%s", mountPoint, path[len([]rune(mountPoint)):]))
			if err == nil {
				version = "2"
			} else {
				version = "1"
				err = nil
			}
		}
	}
	log.Debug().Err(err).Msgf("readVault: getPathMetadataPrefix:\n mountPoint: %s\n version: %s", mountPoint, version)
	return mountPoint, version, err
}

// if VAULT_KV_AUTODETECT envvar is 1, yes or on get mount point and version of the path
// if getPathMetadataInternal fails fall-back to getPathMetadataPrefix
func getPathMetadata(log *zerolog.Logger, vc *vault.Vault, path string) (mountPoint string, version string, err error) {
	kvAutodetect := strings.ToLower(os.Getenv("VAULT_KV_AUTODETECT"))

	version = "1"
	mountPoint = ""
	if kvAutodetect == "1" || kvAutodetect == "on" || kvAutodetect == "yes" {
		mountPoint, version, err = getPathMetadataInternal(log, vc, path) // uses an internal API call, it may fail
		if err != nil {
			mountPoint, version, err = getPathMetadataPrefix(log, vc, path)
			if err != nil {
				return mountPoint, version, err
			}
		}
		if version != "1" && version != "2" {
			return mountPoint, version, fmt.Errorf("only kv versions 1 and 2 are supported, detected version is %s", version)
		}
	}

	return mountPoint, version, nil
}

func readVault(ctx context.Context, source *Source, args ...string) (data []byte, err error) {
	log := zerolog.Ctx(ctx)

	if source.vc == nil {
		source.vc, err = vault.New(source.URL)
		if err != nil {
			return nil, err
		}
		err = source.vc.Login()
		if err != nil {
			return nil, err
		}
	}

	params, p, err := parseDatasourceURLArgs(source.URL, args...)
	if err != nil {
		return nil, err
	}

	mountPoint, version, err := getPathMetadata(log, source.vc, p)
	if err != nil {
		return nil, err
	}

	// write and list available only for kv1
	source.mediaType = jsonMimetype
	switch {
	case len(params) > 0:
		if version != "1" {
			return nil, fmt.Errorf("write is not supported with version %s", version)
		}
		data, err = source.vc.Write(p, params)
	case strings.HasSuffix(p, "/"):
		if version != "1" {
			return nil, fmt.Errorf("list is not supported with version %s", version)
		}
		source.mediaType = jsonArrayMimetype
		data, err = source.vc.List(p)
	default:
		if version == "2" {
			// kv2
			// path -> MOUNTPOINT/data/RELATIVE-PATH
			p = fmt.Sprintf("%sdata/%s", mountPoint, p[len([]rune(mountPoint)):])
		}
		data, err = source.vc.Read(p)
		if err == nil && version == "2" {
			// kv2
			// data -> data["data"]
			var decodedData map[string]any
			json.Unmarshal(data, &decodedData)
			decodedData = decodedData["data"].(map[string]any)
			data, err = json.Marshal(decodedData)
		}
	}
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no value found for path %s", p)
	}

	return data, nil
}
