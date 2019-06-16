package main

import (
	"fmt"
	knapv1 "github.com/bluebosh/knap/pkg/apis/knap/v1alpha1"
	knapclientset "github.com/bluebosh/knap/pkg/client/clientset/versioned"
	"github.com/golang/glog"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/color"
	"html/template"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"strconv"
)

// var kubeconfig = "/Users/jordan/.bluemix/plugins/container-service/clusters/knative_pipeline/kube-config-dal10-knative_pipeline.yml"
var	kubeconfig = "/Users/jordanzt/.bluemix/plugins/container-service/clusters/knative_pipeline/kube-config-dal10-knative_pipeline.yml"

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()

	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
	e.Renderer = renderer

	e.File("/img/knap.png", "img/knap.png")
	e.File("/","index.html")
	e.File("/create","views/create.html")
	e.GET("/list", List)
	e.GET("/createnew", CreateNew)
	e.Logger.Fatal(e.Start(":1323"))
}

func List(c echo.Context) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	knapClient, err := knapclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building knap clientset: %v", err)
	}

	appLst, err := knapClient.KnapV1alpha1().Appengines("default").List(metav1.ListOptions{})
	color.Cyan("%-30s%-20s%-20s%-20s%-20s\n", "Engine Name", "Application Name", "Ready", "Instance", "Domain")
	for i := 0; i < len(appLst.Items); i++ {
		app := appLst.Items[i]
		fmt.Printf("%-30s%-20s%-20s%-20s%-20s\n", app.Name, app.Spec.AppName, app.Status.Ready, fmt.Sprint(app.Status.Instance) + "/" + fmt.Sprint(app.Spec.Size), app.Status.Domain)
	}
	return c.Render(http.StatusOK, "list.html", appLst.Items)
}

func CreateNew(c echo.Context) error {
	r := c.Request()
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	knapClient, err := knapclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building knap clientset: %v", err)
	}


	size, err:= strconv.ParseInt(r.FormValue("size"),10,32)
	size32 := int32(size)
	if err != nil {
		//glog.Fatalf("Error creating application engine: %s", args[0])
		fmt.Println("Error parsing size parameter", err)
	}

	app := &knapv1.Appengine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.FormValue("appName") + "-appengine",
			Namespace: "default",
		},
		Spec:
		knapv1.AppengineSpec{
			AppName: r.FormValue("appName"),
			GitRepo: r.FormValue("gitRepo"),
			GitRevision: r.FormValue("gitRevision"),
			// GitWatch: r.FormValue("gitWatch"),
			Size: size32,
			PipelineTemplate: r.FormValue("template"),
		},
	}
	_, err = knapClient.KnapV1alpha1().Appengines("default").Create(app)

	if err != nil {
		//glog.Fatalf("Error creating application engine: %s", args[0])
		fmt.Println("Error creating application engine", r.FormValue("appName"), err)
	} else {
		fmt.Println("Application engine", r.FormValue("appName"), "is created successfully")
	}

	return c.Render(http.StatusOK, "createDone.html", map[string]interface{}{
		"name": r.FormValue("appName"),
	})


}