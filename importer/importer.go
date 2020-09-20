package importer

import (
	"sync"

	"github.com/zclconf/go-cty/cty"
)

type Importer interface {
	GetImportName(resourceType string, name string, resourceParameters map[string]cty.Value) (string, error)
	ImporterName() string
	Init() error
}

var (
	Importers []Importer
	lock      sync.Mutex
)

func RegisterImporter(importer Importer) {
	lock.Lock()
	defer lock.Unlock()
	Importers = append(Importers, importer)
}
