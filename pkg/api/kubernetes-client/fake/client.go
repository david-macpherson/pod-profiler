package fake

import (
	kubernetesClient "pod_profiler/pkg/api/kubernetes-client"

	"k8s.io/apimachinery/pkg/runtime"

	testclientset "k8s.io/client-go/kubernetes/fake"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	testclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// ClientBuilder builds a fake client.
type ClientBuilder struct {
	clientObjects       []runtime.Object
	clientsetObjects    []runtime.Object
	spsclientsetObjects []runtime.Object
}

// NewClientBuilder returns a new builder to create a fake client.
func NewClientBuilder() *ClientBuilder {
	return &ClientBuilder{}
}

// WithClientRuntimeObjects can be optionally used to initialize this fake client with runtime.Object(s).
func (c *ClientBuilder) WithClientRuntimeObjects(objects ...runtime.Object) *ClientBuilder {
	c.clientObjects = append(c.clientObjects, objects...)
	return c
}

// WithClientsetRuntimeObjects can be optionally used to initialize this fake clientset with runtime.Object(s).
func (c *ClientBuilder) WithClientsetRuntimeObjects(objects ...runtime.Object) *ClientBuilder {
	c.clientsetObjects = append(c.clientsetObjects, objects...)
	return c
}

// WithSPSClientsetRuntimeObjects can be optionally used to initialize this fake clientset with runtime.Object(s).
func (c *ClientBuilder) WithSPSClientsetRuntimeObjects(objects ...runtime.Object) *ClientBuilder {
	c.spsclientsetObjects = append(c.spsclientsetObjects, objects...)
	return c
}

// WithClientsetRuntimeObjects can be optionally used to initialize this fake clientset with runtime.Object(s).
func (c *ClientBuilder) Build() *kubernetesClient.Client {
	return c.newFakeClient().(*kubernetesClient.Client)
}

// Create a fake SPSclient for unit testing
func (c *ClientBuilder) newFakeClient() interface{} {

	// register our application and version crds
	SchemeBuilder := &scheme.Builder{}

	s := runtime.NewScheme()
	clientgoscheme.AddToScheme(s)
	SchemeBuilder.AddToScheme(s)

	spsclient := &kubernetesClient.Client{
		Client:    testclient.NewClientBuilder().WithScheme(s).WithRuntimeObjects(c.clientObjects...).Build(),
		Clientset: testclientset.NewSimpleClientset(c.clientsetObjects...),
	}

	return spsclient
}
