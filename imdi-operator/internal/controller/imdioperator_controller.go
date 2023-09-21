/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"reflect"

	imdiservicemeshv1 "IMDI-Operator/api/v1"
	"io/ioutil"
	"strconv"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	// utilerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ImdiOperatorReconciler reconciles a ImdiOperator object
type ImdiOperatorReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	ProcessedGeneration int64
}

func (r *ImdiOperatorReconciler) UpdateOperator(imdiOperator *imdiservicemeshv1.ImdiOperator,
	log logr.Logger,
	ctx context.Context) error {
	r.ProcessedGeneration = imdiOperator.Generation
	log.Info("Process generation:" + fmt.Sprint(imdiOperator.Generation))
	// imdiOperator.Status = imdiOperatorStatus
	fmt.Println("Update imdi status")
	if updateErr := r.Status().Update(ctx, imdiOperator); updateErr != nil {
		// err := utilerrors.NewAggregate([]error{err, updateErr})
		return updateErr
	}
	return nil
}

func (r *ImdiOperatorReconciler) CheckDuplicatedReconcile(imdiOperator *imdiservicemeshv1.ImdiOperator) bool {
	if imdiOperator.Generation <= r.ProcessedGeneration {
		fmt.Printf("imdiOperator.Generation %i <= r.ProcessedGeneration %i", imdiOperator.Generation, r.ProcessedGeneration)
		return true
	}
	return false
}

func (r *ImdiOperatorReconciler) InstallDevicePluginOperator(imdiOperatorSpec *imdiservicemeshv1.ImdiOperatorSpec,
	imdiOperatorStatus *imdiservicemeshv1.ImdiOperatorStatus,
	log logr.Logger) error {
	log.Info("Install deviceplugin operator")
	fmt.Println("\nbash ./setup_deviceplugin_operator.sh", 
		imdiOperatorSpec.DevicePlugin.Nfd,
		imdiOperatorSpec.DevicePlugin.Nfd_rules,
		imdiOperatorSpec.DevicePlugin.Cert_manager,
		imdiOperatorSpec.DevicePlugin.Device_plugin_operator)
	cmd := exec.Command("bash", "./setup_deviceplugin_operator.sh",
		imdiOperatorSpec.DevicePlugin.Nfd,
		imdiOperatorSpec.DevicePlugin.Nfd_rules,
		imdiOperatorSpec.DevicePlugin.Cert_manager,
		imdiOperatorSpec.DevicePlugin.Device_plugin_operator)
	out, operator_err := cmd.CombinedOutput()
	if operator_err != nil {
		fmt.Printf("setup deviceplugin operator failed \n %s", string(out))
		imdiOperatorStatus.DevicePluginStatus = operator_err.Error()
		return operator_err
	} else {
		imdiOperatorStatus.DevicePluginStatus, imdiOperatorStatus.NfdStatus, imdiOperatorStatus.NfdRuleStatus, imdiOperatorStatus.CertManagerStatus = "Ready", "Ready", "Ready", "Ready"
		return nil
	}
}

func (r *ImdiOperatorReconciler) CryptombCheck(imdiOperatorSpec *imdiservicemeshv1.ImdiOperatorSpec,
	imdiOperatorStatus *imdiservicemeshv1.ImdiOperatorStatus,
	log logr.Logger,
	ctx context.Context) error {

	log.Info("CryptoMB Enable Check")
	for _, gateway := range imdiOperatorSpec.IngressGateways {
		if gateway.CryptoMB != nil {
			imdiOperatorStatus.CryptoMBStatus = "Enable"
			break			
		}
	}

	if imdiOperatorStatus.CryptoMBStatus != "Enable"{
		imdiOperatorStatus.CryptoMBStatus = "NotEnable"
		return nil
	}

	log.Info("CryptoMB is needed, check host support for CryptoMB")
	nodeList := &corev1.NodeList{}
	r.Client.List(ctx, nodeList)
	node := &corev1.Node{}
	nodeName := types.NamespacedName{Name: nodeList.Items[0].Name}
	if err := r.Client.Get(ctx, nodeName, node); err != nil {
		return err
	}
	for key, value := range node.ObjectMeta.GetLabels() {
		if key == "feature.node.kubernetes.io/cpu-cpuid.AVX512F" && value == "true" {
			imdiOperatorStatus.CryptoMBStatus = "Ready"
			break
		}
	}
	if imdiOperatorStatus.CryptoMBStatus != "Ready" {
		imdiOperatorStatus.CryptoMBStatus = "CryptoMB is not supported on this host"
		err := errors.New("CryptoMB is not supported on this host")
		return err
		}
	return nil
}

func (r *ImdiOperatorReconciler) SetupIstio(imdiOperatorSpec *imdiservicemeshv1.ImdiOperatorSpec,
	imdiOperatorStatus *imdiservicemeshv1.ImdiOperatorStatus,
	log logr.Logger) error {
		log.Info("Setup Istio")
		cmd := exec.Command("bash", "setup_istio.sh")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("combined err:\n%s\n", string(out))
			imdiOperatorStatus.IstioStatus = string(out)
			return err
		}
		fmt.Printf("combined out:\n%s\n", string(out))
		imdiOperatorStatus.IstioStatus = "Ready"
		return nil
}

func (r *ImdiOperatorReconciler) DumpIstio(log logr.Logger) error {
	log.Info("Uninstall istio")
	r.ProcessedGeneration = 0
	cmd := exec.Command("bash", "uninstall_istio.sh")
	out, _ := cmd.CombinedOutput()
	// if err != nil {
	// 	fmt.Printf("combined out:\n%s\n", string(out))
	// 	log.Error(err, "uninstall istio failed\n")
	// 	return err
	// }
	fmt.Printf("combined out:\n%s\n", string(out))
	return nil
}

func (r *ImdiOperatorReconciler) DumpQat(imdiOperatorStatus *imdiservicemeshv1.ImdiOperatorStatus,
	log logr.Logger) error {
	if imdiOperatorStatus.QatStatus == "enable" || imdiOperatorStatus.QatStatus == "Qat Number is not enough" {
		log.Info("Uninstall QAT")
		cmd := exec.Command("bash", "uninstall_qat.sh")
		out, _:= cmd.CombinedOutput()
		fmt.Printf("combined out:\n%s\n", string(out))
	}
	return nil
}

func (r *ImdiOperatorReconciler) DumpSgx(imdiOperatorStatus *imdiservicemeshv1.ImdiOperatorStatus,
	log logr.Logger) error {
	if imdiOperatorStatus.CryptoMBStatus == "Enable" || imdiOperatorStatus.CryptoMBStatus == "Ready"{
		log.Info("Uninstall SGX")
		cmd := exec.Command("bash", "uninstall_sgx.sh")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("combined out:\n%s\n", string(out))
			log.Error(err, "Uninstall SGX failed\n")
			return err
		}
		fmt.Printf("combined out:\n%s\n", string(out))
	}
	return nil
}

func (r *ImdiOperatorReconciler) DumpSgxPsw(imdiOperatorStatus *imdiservicemeshv1.ImdiOperatorStatus,
	log logr.Logger) error {
		if imdiOperatorStatus.CryptoMBStatus == "Enable" || imdiOperatorStatus.CryptoMBStatus == "Ready"{	
			log.Info("Uninstall SGX PSW")
			cmd := exec.Command("bash", "uninstall_sgx_psw.sh")
			out, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("combined out:\n%s\n", string(out))
				log.Error(err, "Uninstall SGX PSW failed\n")
				return err
			}
			fmt.Printf("combined out:\n%s\n", string(out))
		}
		return nil
}

func (r *ImdiOperatorReconciler) DumpDevicePlugin(imdiOperatorSpec *imdiservicemeshv1.ImdiOperatorSpec,
	imdiOperatorStatus *imdiservicemeshv1.ImdiOperatorStatus,
	log logr.Logger) error {
		if imdiOperatorStatus.DevicePluginStatus == "Ready" {
			log.Info("Uninstall deviceplugin")
			cmd := exec.Command(
				"bash", "./uninstall_deviceplugin_operator.sh", imdiOperatorSpec.DevicePlugin.Nfd,
				imdiOperatorSpec.DevicePlugin.Nfd_rules, imdiOperatorSpec.DevicePlugin.Cert_manager,
				imdiOperatorSpec.DevicePlugin.Device_plugin_operator)
			out, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("combined out:\n%s\n", string(out))
				log.Error(err, "Uninstall deviceplugin failed\n")
				return err
			}
			fmt.Printf("combined out:\n%s\n", string(out))
		}
		return nil
}

func (r *ImdiOperatorReconciler) SetProxy(imdiOperatorSpec *imdiservicemeshv1.ImdiOperatorSpec,
	imdiOperatorStatus *imdiservicemeshv1.ImdiOperatorStatus,
	log logr.Logger) error {
	log.Info("Setup proxy")
	cmd := exec.Command(
		"bash", "./setup_proxy.sh", imdiOperatorSpec.SetupProxy.Http_proxy,
		imdiOperatorSpec.SetupProxy.Https_proxy, imdiOperatorSpec.SetupProxy.No_proxy)
	cmd.CombinedOutput()
	imdiOperatorStatus.ProxyEnabled = true
	fmt.Printf("\nhttp_proxy: %s\nhttps_proxy: %s\nno_proxy: %s\n", imdiOperatorSpec.SetupProxy.Http_proxy, imdiOperatorSpec.SetupProxy.Https_proxy, imdiOperatorSpec.SetupProxy.No_proxy)
	cmd = exec.Command("bash", "./proxy_test.sh")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(err, "proxy timeout!\n")
		fmt.Println(string(output))
		imdiOperatorStatus.ProxyStatus = "Timeout"
		return err
	}
	imdiOperatorStatus.ProxyStatus = "Ready"
	return nil
}

func (r *ImdiOperatorReconciler) IstioSpecCheck(imdiOperatorSpec imdiservicemeshv1.ImdiOperatorSpec,
	imdiOperatorStatus imdiservicemeshv1.ImdiOperatorStatus,
	log logr.Logger) error {
	imdiSpecvalue := reflect.ValueOf(imdiOperatorSpec)
	if IstioSpecValue := imdiSpecvalue.FieldByName("IstioSpec"); IstioSpecValue.Kind() == reflect.Invalid {
		// TODO: support imid config without IstioSpec Section, using a default istio config
		log.Info("no istio spec, Using default IstioSpec ")
	}
	// fmt.Println("imdiOperatorSpec.IstioSpec: \n", imdiOperatorSpec.IstioSpec)
	if err := ioutil.WriteFile("internal/chart/values.yaml", []byte(imdiOperatorSpec.IstioSpec), 0644); err != nil {
		fmt.Println("An error occurred:", err)
		return err
	}
	imdiOperatorStatus.ConfigParseStatus = "Ready"
	return nil

}

func (r *ImdiOperatorReconciler) GatewaySpecCheck(imdiOperatorSpec *imdiservicemeshv1.ImdiOperatorSpec,
	imdiOperatorStatus *imdiservicemeshv1.ImdiOperatorStatus,
	log logr.Logger) (bool, error) {

	if imdiOperatorSpec.IngressGateways == nil {
		log.Info("no hardware feature")
		imdiOperatorStatus.DevicePluginStatus = "Unenable"
		imdiOperatorStatus.NfdStatus = "Unenable"
		imdiOperatorStatus.NfdRuleStatus = "Unenable"
		imdiOperatorStatus.CertManagerStatus = "Unenable"
		imdiOperatorStatus.QatStatus = "Unenable"
		return false, nil
	}
	log.Info("Hardware feature for Gateway Check")

	data := make(map[string]interface{})
	data["IngressGateways"] = imdiOperatorSpec.IngressGateways
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return true, err
	}
	existingContent, _ := ioutil.ReadFile("internal/chart/values.yaml")
	combinedContent := string(existingContent) + string(yamlBytes)
	ioutil.WriteFile("internal/chart/values.yaml", []byte(combinedContent), 0644)
	return true, nil
}

func (r *ImdiOperatorReconciler) RenderIstioConf(imdiOperatorStatus *imdiservicemeshv1.ImdiOperatorStatus,
	log logr.Logger) error {
	var stdoutBuffer, stderrBuffer bytes.Buffer
	log.Info("Render tempalte with values")
	cmd := exec.Command("helm", "template", "internal/chart", "--debug")
	// out, err := cmd.CombinedOutput()
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer
	if err := cmd.Run(); err != nil {
		fmt.Printf("combined out after render:\n%s\n", string(stderrBuffer.String()))
		imdiOperatorStatus.ConfigParseStatus = string(stderrBuffer.String())
		fmt.Println("render error", err)
		return err
	}
	ioutil.WriteFile("istio.yaml", []byte(stdoutBuffer.String()), 0644)
	// fmt.Println(stdoutBuffer.String())
	imdiOperatorStatus.ConfigParseStatus = "Ready"
	return nil

}

// check if qat enable in configuration and required number
func QatEnable(imdiOperatorSpec *imdiservicemeshv1.ImdiOperatorSpec, log logr.Logger) (bool, int64) {
	log.Info("Qat Enable Check")
	var qat_request_num int64 = 0
	var qat_enable bool = false
	for _, gateway := range imdiOperatorSpec.IngressGateways {
		if gateway.Qat != nil {
			qat_enable = true
			qat_request_num += gateway.Qat.Instance
		}
	}
	// log.Info("Qat enable: %s Number: %i",qat_enable, qat_request_num)
	return qat_enable, qat_request_num
}

func (r *ImdiOperatorReconciler) SgxCheck(imdiOperatorSpec *imdiservicemeshv1.ImdiOperatorSpec,
	imdiOperatorStatus *imdiservicemeshv1.ImdiOperatorStatus,
	log logr.Logger,
	ctx context.Context) error {
	log.Info("SGX enable Check")
	for _, gateway := range imdiOperatorSpec.IngressGateways {
		if gateway.Sgx != nil {
			log.Info("SGX enable: True")
			imdiOperatorStatus.SgxStatus = "Enable"
			break
		}
	}
	if imdiOperatorStatus.SgxStatus != "Enable" {
		imdiOperatorStatus.SgxStatus = "NotEnable"
		return nil
	}
	// Check intel.feature.node.kubernetes.io/sgx=true
	nodeList := &corev1.NodeList{}
	r.Client.List(ctx, nodeList)
	node := &corev1.Node{}
	nodeName := types.NamespacedName{Name: nodeList.Items[0].Name}
	if err := r.Client.Get(ctx, nodeName, node); err != nil {
		return err
	}
	for key, value := range node.ObjectMeta.GetLabels() {
		if key == "intel.feature.node.kubernetes.io/sgx" && value == "true" {
			imdiOperatorStatus.SgxStatus = "Support"
			break
		}
	}
	if imdiOperatorStatus.SgxStatus != "Support" {
		imdiOperatorStatus.SgxStatus = "Sgx is not supported on this host"
		err := errors.New("Sgx is not supported on this host")
		return err
	}

	//install sgx device plugin
	if imdiOperatorStatus.SgxStatus != "Ready" {
		log.Info("Setup SGX")
		cmd := exec.Command("bash", "./setup_sgx.sh")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("combined out error:\n%s\n", string(out))
			log.Error(err, "Setup SGX failed\n")
			imdiOperatorStatus.SgxStatus = string(out)
			return err
		}
		fmt.Printf("SGX combined out:\n%s\n", string(out))
		log.Info("Set SGX status")
		imdiOperatorStatus.SgxStatus = "Ready"
	}

	// install sgx psw
	if imdiOperatorStatus.SgxPSWStatus != "Ready" {
		log.Info("Setup SGX PSW")
		cmd := exec.Command("bash", "./setup_sgx_psw.sh")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("combined out error:\n%s\n", string(out))
			log.Error(err, "Setup SGX PSW failed\n")
			imdiOperatorStatus.SgxPSWStatus = string(out)
			return err
		}
		fmt.Printf("SGX PSW combined out:\n%s\n", string(out))
		imdiOperatorStatus.SgxPSWStatus = "Ready"
	}
	return nil
}

func (r *ImdiOperatorReconciler) QatCheck(imdiOperatorSpec *imdiservicemeshv1.ImdiOperatorSpec,
	imdiOperatorStatus *imdiservicemeshv1.ImdiOperatorStatus,
	log logr.Logger,
	ctx context.Context) error {

	Qatenable, QatNumber := QatEnable(imdiOperatorSpec, log)
	if !Qatenable {
		imdiOperatorStatus.QatStatus = "NotEnable"
		imdiOperatorStatus.QatDeviceNum = 0
		return nil
	}

	// Check intel.feature.node.kubernetes.io/qat=true
	nodeList := &corev1.NodeList{}
	r.Client.List(ctx, nodeList)
	node := &corev1.Node{}
	nodeName := types.NamespacedName{Name: nodeList.Items[0].Name}
	if err := r.Client.Get(ctx, nodeName, node); err != nil {
		return err
	}
	for key, value := range node.ObjectMeta.GetLabels() {
		if key == "intel.feature.node.kubernetes.io/qat" && value == "true" {
			imdiOperatorStatus.QatStatus = "Support"
			break
		}
	}
	if imdiOperatorStatus.QatStatus != "Support" {
		imdiOperatorStatus.QatStatus = "Qat is not supported on this host"
		err := errors.New("Qat is not supported on this host")
		return err
	}
	//install qat device plugin
	if imdiOperatorStatus.QatStatus != "Ready" {
		log.Info("Setup QAT")
		cmd := exec.Command("bash", "./setup_qat.sh")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("[error]:\n%s\n", string(out))
			log.Error(err, "Setup QAT failed\n")
			imdiOperatorStatus.QatStatus = string(out)
			return err
		}
		imdiOperatorStatus.QatStatus = "Support"

	}

	// qat device number calculation
	if QatNumber > 0 {
		log.Info("Qat device calculate")
		nodeList := &corev1.NodeList{}
		if err := r.Client.List(ctx, nodeList); err != nil {
			return err
		}
		node := &corev1.Node{}
		nodeName := types.NamespacedName{Name: nodeList.Items[0].Name}

		if err := r.Client.Get(ctx, nodeName, node); err != nil {
			return err
		}
		for key, value := range node.Status.Allocatable {
			if key == "qat.intel.com/cy" {
				imdiOperatorStatus.QatDeviceNum, _ = strconv.ParseInt(value.String(), 10, 64)
				if imdiOperatorStatus.QatDeviceNum < QatNumber {
					imdiOperatorStatus.QatStatus = strconv.FormatInt(QatNumber,10) + " has exceeded the allocatable qat number "+strconv.FormatInt(imdiOperatorStatus.QatDeviceNum,10)
					err := errors.New("Qat Number is not enough")
					return err
				}
			}
		}
	}
	return nil
}

//+kubebuilder:rbac:groups=imdi-servicemesh.intel.com,resources=imdioperators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=imdi-servicemesh.intel.com,resources=imdioperators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=imdi-servicemesh.intel.com,resources=imdioperators/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ImdiOperator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *ImdiOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	var err error
	var Deviceplugin_enable bool
	var imdiOperator = &imdiservicemeshv1.ImdiOperator{}
	if err = r.Get(ctx, req.NamespacedName, imdiOperator); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Not found. Ignoring.")
			r.ProcessedGeneration = 0
			return ctrl.Result{}, nil
		}
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, fmt.Errorf("unexpected get error: %v", err)
	}

	myFinalizerName := "batch.tutorial.kubebuilder.io/finalizer"
	// skip repeat reconcile
	if r.CheckDuplicatedReconcile(imdiOperator) {
		return ctrl.Result{}, nil
	}

	if imdiOperator.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is create or update, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		log.Info("imdi CRD create logic:")

		imdiOperatorSpec, imdiOperatorStatus := imdiOperator.Spec, imdiOperator.Status
		// parse and write istiospec to values for render

		if err = r.SetProxy(&imdiOperator.Spec, &imdiOperator.Status, log); err != nil {
			// return ctrl.Result{}, err
			goto FINISH
		}

		if err = r.IstioSpecCheck(imdiOperatorSpec, imdiOperatorStatus, log); err != nil {
			return ctrl.Result{}, err
		}

		// parse and write gateway(hardware) section for render
		Deviceplugin_enable, err = r.GatewaySpecCheck(&imdiOperator.Spec, &imdiOperator.Status, log)
		if err != nil {
			goto FINISH
		}

		// render tempalte with values
		if err = r.RenderIstioConf(&imdiOperator.Status, log); err != nil {
			return ctrl.Result{}, err
		}

		// don't add finalizer real refer resource(device plugin/istio) created
		if !controllerutil.ContainsFinalizer(imdiOperator, myFinalizerName) {
			controllerutil.AddFinalizer(imdiOperator, myFinalizerName)
			if err := r.Update(ctx, imdiOperator); err != nil {
				return ctrl.Result{}, err
			}
		}

		// install deviceplugin operator, check
		if Deviceplugin_enable && imdiOperatorStatus.DevicePluginStatus != "Ready" {
			if err = r.InstallDevicePluginOperator(&imdiOperator.Spec, &imdiOperator.Status, log); err != nil {
				goto FINISH
			}
		}

		// check CryptoMB enable and avaliablity
		if err = r.CryptombCheck(&imdiOperator.Spec, &imdiOperator.Status, log, ctx); err != nil {
			goto FINISH
		}

		if err = r.QatCheck(&imdiOperator.Spec, &imdiOperator.Status, log, ctx); err != nil {
			goto FINISH
		}

		if err = r.SgxCheck(&imdiOperator.Spec, &imdiOperator.Status, log, ctx); err != nil {
			goto FINISH
		}

		// setup Istio
		if err = r.SetupIstio(&imdiOperator.Spec, &imdiOperator.Status, log); err != nil {
			return ctrl.Result{}, err
		}

		// update status
		FINISH:
		 	r.UpdateOperator(imdiOperator, log, ctx)
			return ctrl.Result{}, nil
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(imdiOperator, myFinalizerName) {
			// imdiOperatorSpec, imdiOperatorStatus := imdiOperator.Spec, imdiOperator.Status
			if !controllerutil.ContainsFinalizer(imdiOperator, myFinalizerName) {
				return ctrl.Result{}, nil
			}
			// setup proxy in case the proxy non-exist due to operator delete
			if err := r.SetProxy(&imdiOperator.Spec, &imdiOperator.Status, log); err != nil {
				return ctrl.Result{}, err
			}

			// uninstall istio
			if err := r.DumpIstio(log); err != nil {
				return ctrl.Result{}, err
			}

			//uninstall qat
			if err := r.DumpQat(&imdiOperator.Status,log); err != nil {
				return ctrl.Result{}, err
			}

			//uninstall sgx
			if err := r.DumpSgx(&imdiOperator.Status,log); err != nil {
				return ctrl.Result{}, err
			}

			//uninstall sgx psw
			if err := r.DumpSgxPsw(&imdiOperator.Status,log); err != nil {
				return ctrl.Result{}, err
			}

			if err := r.DumpDevicePlugin(&imdiOperator.Spec,&imdiOperator.Status, log); err != nil {
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(imdiOperator, myFinalizerName)
			if err := r.Update(ctx, imdiOperator); err != nil {
				return ctrl.Result{}, err
			}
		}
			return ctrl.Result{}, nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ImdiOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&imdiservicemeshv1.ImdiOperator{}).
		Complete(r)
}
