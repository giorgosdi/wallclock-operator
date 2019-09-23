package timezones

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	clockv1 "github.com/giorgosdi/wallclock-operator/pkg/apis/clock/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_timezones")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Timezones Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTimezones{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("timezones-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Timezones
	err = c.Watch(&source.Kind{Type: &clockv1.Timezones{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Timezones
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &clockv1.Timezones{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileTimezones implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileTimezones{}

// ReconcileTimezones reconciles a Timezones object
type ReconcileTimezones struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Timezones object and makes changes based on the state read
// and what is in the Timezones.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileTimezones) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Timezones")

	var rec reconcile.Result
	var e error

	// Fetch the Timezones instance
	instance := &clockv1.Timezones{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	found := &clockv1.Wallclock{}
	// Define a new wallclock object
	wallclocks := createWallclock(instance)

	for _, wc := range wallclocks {
		// Set Timezones instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, wc, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if this wallclock already exists

		err = r.client.Get(context.TODO(), types.NamespacedName{Name: wc.Name, Namespace: wc.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Wallclock", "Wallclock.Namespace", wc.Namespace, "Wallclock.Name", wc.Name)
			reqLogger.Info("Update wallclock to ", "Timezone.Spec", instance.Spec)
			reqLogger.Info("1st timezone ", "Timezone.Spec.clock", instance.Spec.Clocks[0])
			err = r.client.Create(context.TODO(), wc)
			if err != nil {
				rec = reconcile.Result{}
				e = err
			}
			// Update the wallclock time, :
			wc.Status.Time = convertTime(wc.Spec.Timezone)
			if wc.Status.Time == "" {
				return reconcile.Result{}, err
			}
			err = r.client.Status().Update(context.Background(), wc)
			if err != nil {
				rec = reconcile.Result{}
				e = err
			}
			reqLogger.WithValues("wallclock status", wc.Status.Time)
			rec = reconcile.Result{}
			e = nil
			// Wallclock created successfully - don't requeue
		} else if err != nil {
			rec = reconcile.Result{}
			e = err
		}

		// Wallclock already exists - don't requeue
		reqLogger.Info("Skip reconcile: Wallclock already exists", "Wallclock.Namespace", found.Namespace, "Wallclock.Name", found.Name)
		rec = reconcile.Result{}
		e = err
	}
	return rec, e
}

func createWallclock(cr *clockv1.Timezones) []*clockv1.Wallclock {
	var wallclocks []*clockv1.Wallclock
	labels := map[string]string{
		"app": cr.Name,
	}
	for _, v := range cr.Spec.Clocks {
		wallclocks = append(wallclocks, &clockv1.Wallclock{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cr.Name + "-" + strings.ToLower(v),
				Namespace: cr.Namespace,
				Labels:    labels,
			},
			Spec: clockv1.WallclockSpec{
				Timezone: v,
			},
		})
	}
	return wallclocks
}

func convertTime(ctime string) string {
	log.Info("COVERTING TO", "Wallclock.timezone", ctime)
	var stringToReturn string
	loc, err := time.LoadLocation(ctime)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	getTz()
	for _, tz := range tzdata {
		if tz == ctime {
			log.Info("Timezone is", "valid", ctime)
			t := time.Now().In(loc)
			stringToReturn = fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
			log.Info("CONVERTED TO ", "TZ", ctime)
		}
	}
	return stringToReturn
}

var zoneDirs = []string{
	"/usr/share/zoneinfo/",
}

var zoneDir string
var tzdata []string

func getTz() {
	for _, zoneDir = range zoneDirs {
		readFile("")
	}
}

func readFile(path string) {

	files, _ := ioutil.ReadDir(zoneDir + path)
	for _, f := range files {
		if f.Name() != strings.ToUpper(f.Name()[:1])+f.Name()[1:] {
			continue
		}
		if f.IsDir() {
			readFile(path + "/" + f.Name())
		} else {
			tzdata = append(tzdata, (path + "/" + f.Name())[1:])
		}
	}
}
