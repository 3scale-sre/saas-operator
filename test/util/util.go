package util

import (
	"context"
	"time"

	"github.com/goombaio/namegenerator"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MustParseRFC3339(in string) time.Time {
	t, err := time.Parse(time.RFC3339, in)
	if err != nil {
		panic(err)
	}

	return t
}

func CreateNamespace(nameGenerator namegenerator.Generator, c client.Client,
	timeout, poll time.Duration) string {
	var namespace string

	Eventually(func() error {
		namespace = "test-ns-" + nameGenerator.Generate()

		testNamespace := &corev1.Namespace{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		}

		if err := c.Create(context.Background(), testNamespace); err != nil {
			return err
		}

		return nil
	}, timeout, poll).ShouldNot(HaveOccurred())

	n := &corev1.Namespace{}

	Eventually(func() error {
		return c.Get(context.Background(), types.NamespacedName{Name: namespace}, n)
	}, timeout, poll).ShouldNot(HaveOccurred())

	return namespace
}
