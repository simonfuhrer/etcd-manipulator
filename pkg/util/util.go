package util

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/coreos/etcd/pkg/transport"
	"go.etcd.io/etcd/clientv3"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/etcd3"
	"k8s.io/apiserver/pkg/storage/value"
	//"go.uber.org/zap"
	//"github.com/k0kubun/pp"
)

const (
	pvkey = "/registry/persistentvolumes"
)

type Client struct {
	cl *clientv3.Client
}

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

func init() {
	metav1.AddToGroupVersion(scheme, metav1.SchemeGroupVersion)
	utilruntime.Must(v1.AddToScheme(scheme))
}

func InitClient(endpoints []string, tls transport.TLSInfo) (client *Client, err error) {
	//clientv3.SetLogger(grpclog.NewLoggerV2( ioutil.Discard, ioutil.Discard, ioutil.Discard))
	var myclient = &Client{}

	tlsConfig, err := tls.ClientConfig()
	if err != nil {
		return myclient, err
	}

	config := clientv3.Config{
		Endpoints:   endpoints,
		TLS:         tlsConfig,
		DialTimeout: 5 * time.Second,
	}

	cl, err := clientv3.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to connect to etcd: %v\n", err)
		return myclient, err
	}
	return &Client{cl: cl}, nil
}

func (c *Client) DumpPVs() error {
	codec := codecs.LegacyCodec(v1.SchemeGroupVersion)
	store := etcd3.New(c.cl, codec, "", value.IdentityTransformer, false)
	ctx := context.Background()
	out := &v1.PersistentVolumeList{}
	err := store.List(ctx, pvkey, "", storage.Everything, out)
	if err != nil {
		return err
	}

	var jsonData []byte
	jsonData, err = json.Marshal(out.Items)
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}

func (c *Client) ModifyPVs(name string, newname string, dryrun bool) error {
	codec := codecs.LegacyCodec(v1.SchemeGroupVersion)
	store := etcd3.New(c.cl, codec, "", value.IdentityTransformer, false)
	ctx := context.Background()
	out := &v1.PersistentVolumeList{}
	err := store.List(ctx, pvkey, "", storage.Everything, out)
	if err != nil {
		return err
	}
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)
	defer w.Flush()
	fmt.Fprintf(w, "\n %s\t%s\t%s\t", "key", "name", "newname")
	fmt.Fprintf(w, "\n %s\t%s\t%s\t", "----", "----", "----")

	for _, pv := range out.Items {
		outpv := &v1.PersistentVolume{}
		pvpath := fmt.Sprintf("%s/%s", pvkey, pv.ObjectMeta.Name)
		oldvolumePath := pv.Spec.PersistentVolumeSource.VsphereVolume.VolumePath
		newvolumePath := strings.Replace(oldvolumePath, name, newname, -1)
		if strings.EqualFold(oldvolumePath, newvolumePath) != true {
			if dryrun != true {
				err = store.GuaranteedUpdate(ctx, pvpath, outpv, true, nil,
					storage.SimpleUpdate(func(obj runtime.Object) (runtime.Object, error) {
						pvneu := obj.(*v1.PersistentVolume)
						pvneu.Spec.PersistentVolumeSource.VsphereVolume.VolumePath = newvolumePath
						return obj, nil
					}))
				if err != nil {
					return err
				}
			}
		}
		fmt.Fprintf(w, "\n %s\t%s\t%s\t", pvpath, oldvolumePath, newvolumePath)
	}

	return nil
}
