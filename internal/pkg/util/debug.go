package util

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/3scale-sre/basereconciler/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//nolint:errchkjson
func ResourceDump[T client.Object](resource *resource.Template[T]) {
	obj, _ := resource.Build(context.TODO(), nil, nil)
	j, _ := json.Marshal(obj)
	fmt.Println(string(j))
}
