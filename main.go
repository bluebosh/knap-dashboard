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
	"os"
	"strconv"
)

var kubeconfig string

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
	kubeconfig = os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		glog.Fatalf("Cannot get kubeconfig from: %v", "KUBECONFIG")
	}

	e := echo.New()

	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
	e.Renderer = renderer

	e.File("/img/knap.png", "img/knap.png")
	e.File("/img/Running.png", "img/running.png")
	e.File("/img/Pending.png", "img/pending.png")
	e.File("/img/Deployed.png", "img/deployed.png")
	e.File("/img/Fail.png", "img/fail.png")
	e.File("/","index.html")
	e.File("/create","views/create.html")
	e.File("/templates", "views/templates.html")
	e.GET("/get", Get)
	e.GET("/list", List)
	e.GET("/spaces", Spaces)
	e.GET("/services", Services)
	e.GET("/createnew", CreateNew)
	e.GET("/edit", Edit)
	e.GET("/getedit", GetEdit)
	e.GET("/delete", Delete)
	e.Logger.Fatal(e.Start(":1323"))
}

func Get(c echo.Context) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	knapClient, err := knapclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building knap clientset: %v", err)
	}

	app, err := knapClient.KnapV1alpha1().Appengines("default").Get(c.Request().FormValue("name"), metav1.GetOptions{})
	if err != nil {
		glog.Fatalf("Error getting appengine: %v", c.Request().FormValue("name"))
	}
	fmt.Print("Get appengine", "appengine name", app.Name)
	return c.Render(http.StatusOK, "get.html", app)
}

func Delete(c echo.Context) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	knapClient, err := knapclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building knap clientset: %v", err)
	}

	app, err := knapClient.KnapV1alpha1().Appengines("default").Get(c.Request().FormValue("name"), metav1.GetOptions{})
	if err != nil {
		glog.Fatalf("Error getting appengine: %v", c.Request().FormValue("name"))
	}
	fmt.Print("Get appengine", "appengine name", app.Spec.AppName)

	err = knapClient.KnapV1alpha1().Appengines("default").Delete(app.Name, &metav1.DeleteOptions{})
	if err != nil {
		glog.Fatalf("Error deleting appengine: %v", c.Request().FormValue("name"))
	}
	fmt.Print("Delete appengine", "appengine name", app.Name)
	return c.Render(http.StatusOK, "deleteDone.html", app.Spec.AppName)
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

func Spaces(c echo.Context) error {
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
	return c.Render(http.StatusOK, "spaces.html", appLst.Items)
}

func Services(c echo.Context) error {
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
	return c.Render(http.StatusOK, "services.html", appLst.Items)
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

func Edit(c echo.Context) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	knapClient, err := knapclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building knap clientset: %v", err)
	}

	app, err := knapClient.KnapV1alpha1().Appengines("default").Get(c.Request().FormValue("name"), metav1.GetOptions{})
	if err != nil {
		glog.Fatalf("Error getting appengine: %v", c.Request().FormValue("name"))
	}
	fmt.Print("Get appengine", "appengine name", app.Name)
	return c.Render(http.StatusOK, "edit.html", app)
}

func GetEdit(c echo.Context) error {
	r := c.Request()
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	knapClient, err := knapclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building knap clientset: %v", err)
	}

	app, err := knapClient.KnapV1alpha1().Appengines("default").Get(r.FormValue("appName") + "-appengine", metav1.GetOptions{})
	if err != nil {
		glog.Fatalf("Error getting appengine: %v", r.FormValue("appName") + "-appengine")
	}

	size, err:= strconv.ParseInt(r.FormValue("size"),10,32)
	size32 := int32(size)
	if err != nil {
		//glog.Fatalf("Error creating application engine: %s", args[0])
		fmt.Println("Error parsing size parameter", err)
	}

	app.Spec.GitRevision = r.FormValue("gitRevision")
	//app.Spec.GitWatch = r.FormValue("gitWatch")
	app.Spec.Size = size32
	app.Spec.PipelineTemplate = r.FormValue("template")

	_, err = knapClient.KnapV1alpha1().Appengines("default").Update(app)

	if err != nil {
		//glog.Fatalf("Error creating application engine: %s", args[0])
		fmt.Println("Error updating application engine", r.FormValue("appName"), err)
	} else {
		fmt.Println("Application engine", r.FormValue("appName"), "is updated successfully")
	}

	return c.Render(http.StatusOK, "editDone.html", map[string]interface{}{
		"name": r.FormValue("appName"),
	})
}
