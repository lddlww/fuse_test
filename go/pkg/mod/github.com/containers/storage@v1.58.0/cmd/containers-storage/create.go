package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/containers/storage"
	graphdriver "github.com/containers/storage/drivers"
	"github.com/containers/storage/internal/opts"
	"github.com/containers/storage/pkg/idtools"
	"github.com/containers/storage/pkg/mflag"
	"github.com/containers/storage/pkg/stringid"
	"github.com/containers/storage/types"
	digest "github.com/opencontainers/go-digest"
)

var (
	paramMountOptions = ""
	paramMountLabel   = ""
	paramNames        = []string{}
	paramID           = ""
	paramDigest       = ""
	paramLayer        = ""
	paramMetadata     = ""
	paramMetadataFile = ""
	paramCreateRO     = false
	paramHostUIDMap   = false
	paramHostGIDMap   = false
	paramUIDMap       = ""
	paramGIDMap       = ""
	paramSubUIDMap    = ""
	paramSubGIDMap    = ""
	paramReadOnly     = false
	paramVolatile     = false
)

func paramIDMapping() (*types.IDMappingOptions, error) {
	options := types.IDMappingOptions{
		HostUIDMapping: paramHostUIDMap,
		HostGIDMapping: paramHostGIDMap,
	}
	if paramHostUIDMap && (len(paramUIDMap) > 0 || paramSubUIDMap != "") {
		return nil, fmt.Errorf("host uid map specified along with UID mapping")
	}
	if paramHostGIDMap && (len(paramGIDMap) > 0 || paramSubGIDMap != "") {
		return nil, fmt.Errorf("host gid map specified along with GID mapping")
	}
	if paramSubGIDMap == "" && paramSubUIDMap != "" {
		paramSubGIDMap = paramSubUIDMap
	}
	if paramSubUIDMap == "" && paramSubGIDMap != "" {
		paramSubUIDMap = paramSubGIDMap
	}
	if paramSubUIDMap != "" && paramSubGIDMap != "" {
		mappings, err := idtools.NewIDMappings(paramSubUIDMap, paramSubGIDMap)
		if err != nil {
			return nil, err
		}
		options.UIDMap = mappings.UIDs()
		options.GIDMap = mappings.GIDs()
	}
	parsedUIDMap, err := idtools.ParseIDMap([]string{paramUIDMap}, "uid")
	if err != nil {
		return nil, err
	}
	parsedGIDMap, err := idtools.ParseIDMap([]string{paramGIDMap}, "gid")
	if err != nil {
		return nil, err
	}
	options.UIDMap = append(options.UIDMap, parsedUIDMap...)
	options.GIDMap = append(options.GIDMap, parsedGIDMap...)
	return &options, nil
}

func createStorageLayer(flags *mflag.FlagSet, action string, m storage.Store, args []string) (int, error) {
	parent := ""
	if len(args) > 0 {
		parent = args[0]
	}
	mappings, err := paramIDMapping()
	if err != nil {
		return 1, err
	}
	driver, err := m.GraphDriver()
	if err != nil {
		return 1, err
	}
	opts := graphdriver.CreateOpts{
		MountLabel: paramMountLabel,
		IDMappings: idtools.NewIDMappingsFromMaps(mappings.UIDMap, mappings.GIDMap),
	}
	if paramID == "" {
		paramID = stringid.GenerateNonCryptoID()
	}
	if paramCreateRO {
		if err := driver.Create(paramID, parent, &opts); err != nil {
			return 1, err
		}
	} else {
		if err := driver.CreateReadWrite(paramID, parent, &opts); err != nil {
			return 1, err
		}
	}
	if jsonOutput {
		if err := json.NewEncoder(os.Stdout).Encode(paramID); err != nil {
			return 1, err
		}
	} else {
		fmt.Printf("%s\n", paramID)
	}
	return 0, nil
}

func createLayer(flags *mflag.FlagSet, action string, m storage.Store, args []string) (int, error) {
	parent := ""
	if len(args) > 0 {
		parent = args[0]
	}
	mappings, err := paramIDMapping()
	if err != nil {
		return 1, err
	}
	options := &storage.LayerOptions{IDMappingOptions: *mappings}
	layer, err := m.CreateLayer(paramID, parent, paramNames, paramMountLabel, !paramCreateRO, options)
	if err != nil {
		return 1, err
	}
	if jsonOutput {
		return outputJSON(layer)
	}
	fmt.Printf("%s\n", layer.ID)
	for _, name := range layer.Names {
		fmt.Printf("\t%s\n", name)
	}
	return 0, nil
}

func importLayer(flags *mflag.FlagSet, action string, m storage.Store, args []string) (int, error) {
	parent := ""
	if len(args) > 0 {
		parent = args[0]
	}
	diffStream := io.Reader(os.Stdin)
	if applyDiffFile != "" {
		f, err := os.Open(applyDiffFile)
		if err != nil {
			return 1, err
		}
		diffStream = f
		defer f.Close()
	}
	mappings, err := paramIDMapping()
	if err != nil {
		return 1, err
	}
	options := &storage.LayerOptions{IDMappingOptions: *mappings}
	layer, _, err := m.PutLayer(paramID, parent, paramNames, paramMountLabel, !paramCreateRO, options, diffStream)
	if err != nil {
		return 1, err
	}
	if jsonOutput {
		return outputJSON(layer)
	}
	fmt.Printf("%s\n", layer.ID)
	for _, name := range layer.Names {
		fmt.Printf("\t%s\n", name)
	}
	return 0, nil
}

func createImage(flags *mflag.FlagSet, action string, m storage.Store, args []string) (int, error) {
	if paramMetadataFile != "" {
		f, err := os.Open(paramMetadataFile)
		if err != nil {
			return 1, err
		}
		b, err := io.ReadAll(f)
		if err != nil {
			return 1, err
		}
		paramMetadata = string(b)
	}
	layer := ""
	if len(args) > 0 {
		layer = args[0]
	}
	imageOptions := &storage.ImageOptions{
		Digest: digest.Digest(paramDigest),
	}
	image, err := m.CreateImage(paramID, paramNames, layer, paramMetadata, imageOptions)
	if err != nil {
		return 1, err
	}
	if jsonOutput {
		return outputJSON(image)
	}
	fmt.Printf("%s\n", image.ID)
	for _, name := range image.Names {
		fmt.Printf("\t%s\n", name)
	}
	return 0, nil
}

func createContainer(flags *mflag.FlagSet, action string, m storage.Store, args []string) (int, error) {
	if paramMetadataFile != "" {
		f, err := os.Open(paramMetadataFile)
		if err != nil {
			return 1, err
		}
		b, err := io.ReadAll(f)
		if err != nil {
			return 1, err
		}
		paramMetadata = string(b)
	}
	mappings, err := paramIDMapping()
	if err != nil {
		return 1, err
	}
	options := &storage.ContainerOptions{IDMappingOptions: *mappings, Volatile: paramVolatile}
	image := args[0]
	container, err := m.CreateContainer(paramID, paramNames, image, paramLayer, paramMetadata, options)
	if err != nil {
		return 1, err
	}
	if jsonOutput {
		return outputJSON(container)
	}
	fmt.Printf("%s\n", container.ID)
	for _, name := range container.Names {
		fmt.Printf("\t%s\n", name)
	}
	return 0, nil
}

func init() {
	commands = append(commands, command{
		names:       []string{"create-storage-layer"},
		optionsHelp: "[options [...]] [parentLayerNameOrID]",
		usage:       "Create a new layer only in the storage driver",
		minArgs:     0,
		maxArgs:     1,
		action:      createStorageLayer,
		addFlags: func(flags *mflag.FlagSet, cmd *command) {
			flags.StringVar(&paramMountLabel, []string{"-label", "l"}, "", "Mount Label")
			flags.StringVar(&paramID, []string{"-id", "i"}, "", "Layer ID")
			flags.BoolVar(&paramCreateRO, []string{"-readonly", "r"}, false, "Mark as read-only")
			flags.BoolVar(&jsonOutput, []string{"-json", "j"}, jsonOutput, "Prefer JSON output")
			flags.BoolVar(&paramHostUIDMap, []string{"-hostuidmap"}, paramHostUIDMap, "Force host UID map")
			flags.BoolVar(&paramHostGIDMap, []string{"-hostgidmap"}, paramHostGIDMap, "Force host GID map")
			flags.StringVar(&paramUIDMap, []string{"-uidmap"}, "", "UID map")
			flags.StringVar(&paramGIDMap, []string{"-gidmap"}, "", "GID map")
			flags.StringVar(&paramSubUIDMap, []string{"-subuidmap"}, "", "subuid UID map for a user")
			flags.StringVar(&paramSubGIDMap, []string{"-subgidmap"}, "", "subgid GID map for a group")
		},
	})
	commands = append(commands, command{
		names:       []string{"create-layer", "createlayer"},
		optionsHelp: "[options [...]] [parentLayerNameOrID]",
		usage:       "Create a new layer",
		minArgs:     0,
		maxArgs:     1,
		action:      createLayer,
		addFlags: func(flags *mflag.FlagSet, cmd *command) {
			flags.StringVar(&paramMountLabel, []string{"-label", "l"}, "", "Mount Label")
			flags.Var(opts.NewListOptsRef(&paramNames, nil), []string{"-name", "n"}, "Layer name")
			flags.StringVar(&paramID, []string{"-id", "i"}, "", "Layer ID")
			flags.BoolVar(&paramCreateRO, []string{"-readonly", "r"}, false, "Mark as read-only")
			flags.BoolVar(&jsonOutput, []string{"-json", "j"}, jsonOutput, "Prefer JSON output")
			flags.BoolVar(&paramHostUIDMap, []string{"-hostuidmap"}, paramHostUIDMap, "Force host UID map")
			flags.BoolVar(&paramHostGIDMap, []string{"-hostgidmap"}, paramHostGIDMap, "Force host GID map")
			flags.StringVar(&paramUIDMap, []string{"-uidmap"}, "", "UID map")
			flags.StringVar(&paramGIDMap, []string{"-gidmap"}, "", "GID map")
			flags.StringVar(&paramSubUIDMap, []string{"-subuidmap"}, "", "subuid UID map for a user")
			flags.StringVar(&paramSubGIDMap, []string{"-subgidmap"}, "", "subgid GID map for a group")
		},
	})
	commands = append(commands, command{
		names:       []string{"import-layer", "importlayer"},
		optionsHelp: "[options [...]] [parentLayerNameOrID]",
		usage:       "Import a new layer",
		minArgs:     0,
		maxArgs:     1,
		action:      importLayer,
		addFlags: func(flags *mflag.FlagSet, cmd *command) {
			flags.StringVar(&paramMountLabel, []string{"-label", "l"}, "", "Mount Label")
			flags.Var(opts.NewListOptsRef(&paramNames, nil), []string{"-name", "n"}, "Layer name")
			flags.StringVar(&paramID, []string{"-id", "i"}, "", "Layer ID")
			flags.BoolVar(&paramCreateRO, []string{"-readonly", "r"}, false, "Mark as read-only")
			flags.BoolVar(&jsonOutput, []string{"-json", "j"}, jsonOutput, "Prefer JSON output")
			flags.StringVar(&applyDiffFile, []string{"-file", "f"}, "", "Read from file instead of stdin")
			flags.BoolVar(&paramHostUIDMap, []string{"-hostuidmap"}, paramHostUIDMap, "Force host UID map")
			flags.BoolVar(&paramHostGIDMap, []string{"-hostgidmap"}, paramHostGIDMap, "Force host GID map")
			flags.StringVar(&paramUIDMap, []string{"-uidmap"}, "", "UID map")
			flags.StringVar(&paramGIDMap, []string{"-gidmap"}, "", "GID map")
			flags.StringVar(&paramSubUIDMap, []string{"-subuidmap"}, "", "subuid UID map for a user")
			flags.StringVar(&paramSubGIDMap, []string{"-subgidmap"}, "", "subgid GID map for a group")
		},
	})
	commands = append(commands, command{
		names:       []string{"create-image", "createimage"},
		optionsHelp: "[options [...]] topLayerNameOrID",
		usage:       "Create a new image using layers",
		minArgs:     0,
		maxArgs:     1,
		action:      createImage,
		addFlags: func(flags *mflag.FlagSet, cmd *command) {
			flags.Var(opts.NewListOptsRef(&paramNames, nil), []string{"-name", "n"}, "Image name")
			flags.StringVar(&paramID, []string{"-id", "i"}, "", "Image ID")
			flags.StringVar(&paramDigest, []string{"-digest", "d"}, "", "Image Digest")
			flags.StringVar(&paramMetadata, []string{"-metadata", "m"}, "", "Metadata")
			flags.StringVar(&paramMetadataFile, []string{"-metadata-file", "f"}, "", "Metadata File")
			flags.BoolVar(&jsonOutput, []string{"-json", "j"}, jsonOutput, "Prefer JSON output")
		},
	})
	commands = append(commands, command{
		names:       []string{"create-container", "createcontainer"},
		optionsHelp: "[options [...]] parentImageNameOrID",
		usage:       "Create a new container from an image",
		minArgs:     1,
		maxArgs:     1,
		action:      createContainer,
		addFlags: func(flags *mflag.FlagSet, cmd *command) {
			flags.Var(opts.NewListOptsRef(&paramNames, nil), []string{"-name", "n"}, "Container name")
			flags.BoolVar(&paramVolatile, []string{"-volatile"}, false, "Mark as volatile")
			flags.StringVar(&paramID, []string{"-id", "i"}, "", "Container ID")
			flags.StringVar(&paramMetadata, []string{"-metadata", "m"}, "", "Metadata")
			flags.StringVar(&paramMetadataFile, []string{"-metadata-file", "f"}, "", "Metadata File")
			flags.BoolVar(&jsonOutput, []string{"-json", "j"}, jsonOutput, "Prefer JSON output")
			flags.BoolVar(&paramHostUIDMap, []string{"-hostuidmap"}, paramHostUIDMap, "Force host UID map")
			flags.BoolVar(&paramHostGIDMap, []string{"-hostgidmap"}, paramHostGIDMap, "Force host GID map")
			flags.StringVar(&paramUIDMap, []string{"-uidmap"}, "", "UID map")
			flags.StringVar(&paramGIDMap, []string{"-gidmap"}, "", "GID map")
			flags.StringVar(&paramSubUIDMap, []string{"-subuidmap"}, "", "subuid UID map for a user")
			flags.StringVar(&paramSubGIDMap, []string{"-subgidmap"}, "", "subgid GID map for a group")
		},
	})
}
