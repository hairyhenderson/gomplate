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

func stripMountPoint(path string, mountPoint string) string {
	return path[len([]rune(mountPoint)):]
}

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

func getPathMetadataPrefix(log *zerolog.Logger, vc *vault.Vault, path string) (mountPoint string, version string, err error) {
	mountPoint = ""
	version = ""
	jsonData, err := vc.Read("/sys/mounts")
	if err == nil {
		var decodedData map[string]any
		json.Unmarshal(jsonData, &decodedData)
		x := 0
		y := 0
		for mp := range decodedData {
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
		} else {
			_, err = vc.Read(fmt.Sprintf("%smetadata/%s", mountPoint, stripMountPoint(path, mountPoint)))
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

	kv_autodetect := strings.ToLower(os.Getenv("VAULT_KV_AUTODETECT"))

	version := "1"
	mountPoint := ""
	if kv_autodetect == "1" || kv_autodetect == "on" || kv_autodetect == "yes" {
		mountPoint, version, err = getPathMetadataInternal(log, source.vc, p) // uses an internal API call, it may fail
		if err != nil {
			mountPoint, version, err = getPathMetadataPrefix(log, source.vc, p)
			if err != nil {
				return nil, err
			}
		}
		if version != "1" && version != "2" {
			return nil, fmt.Errorf("only kv versions 1 and 2 are supported, detected version is %s", version)
		}
	}

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
			p = fmt.Sprintf("%sdata/%s", mountPoint, stripMountPoint(p, mountPoint))
		}
		data, err = source.vc.Read(p)
		if err == nil && version == "2" {
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
