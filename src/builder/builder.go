package builder

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/francisbouvier/pipes/src/engine/docker"
)

type categorization struct {
	execType        string
	baseDockerImage string
}

var (
	pythonCategory       = categorization{execType: "python", baseDockerImage: "python"}
	rubyCategory         = categorization{execType: "ruby", baseDockerImage: "ruby"}
	simpleBinaryCategory = categorization{execType: "binary", baseDockerImage: "microbox/scratch"}
)

func check(e error) {
	if e != nil {
		fmt.Printf("%s\n", e)
		panic(e)
	}
}

func BuildDockerImagesFromExec(exec_paths []string) (err error){ //execPath_category_map *map[string]categorization) {
	execs_map := AssociateExecWithType(exec_paths)
	fmt.Println()

	// Iterating through the map
	for execOriginalPath, category := range execs_map {
		tmp_dir_path, new_exec_path, exec_file_name := SetTempDirectory(execOriginalPath)
		imageName := CreateDockerfile(tmp_dir_path, new_exec_path, exec_file_name, category)
		err = DockerBuild(tmp_dir_path, imageName)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println()
	}
	fmt.Printf("Docker images successfully built...\n")
	return
}

// Create a Dockerfile per executable passed through CLI
func AssociateExecWithType(exec_paths []string) (execPath_category_map map[string]categorization) {
	// var execPath_type_map map[string]string
	execPath_category_map = make(map[string]categorization)
	for _, exec_path := range exec_paths {
		switch {
		case strings.HasSuffix(exec_path, ".py"):
			execPath_category_map[exec_path] = pythonCategory
			// fmt.Printf("file %s is a python file\n", exec_path)
		case strings.HasSuffix(exec_path, ".rb"):
			execPath_category_map[exec_path] = rubyCategory
			// fmt.Printf("file %s is a ruby file\n", exec_path)
		default:
			execPath_category_map[exec_path] = simpleBinaryCategory
			// fmt.Printf("file %s is a binary file\n", exec_path)
		}
		fmt.Printf("File %s is a %s file, and will be dockerized from the base image '%s'\n", exec_path, execPath_category_map[exec_path].execType, execPath_category_map[exec_path].baseDockerImage)
	}

	return
}

// Set a temp directory and cp the exec in it
func SetTempDirectory(old_exec_path string) (tmp_dir_path, new_exec_path, exec_file_name string) {
	// mkdir a tmp dir
	/// generate random dir name from wd/tmp
	wd, _ := os.Getwd()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	random_nbr := r.Int63()
	tmp_dir_path = fmt.Sprintf("%s/tmp/%d", wd, random_nbr)
	// fmt.Printf("tmp_dir_path=%s\n", tmp_dir_path)
	/// mkdir cmd
	err := os.MkdirAll(tmp_dir_path, 0777)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	// cp exec in that tmp dir
	/// get exec name (and not path) to create new path for it
	exec_file_name_array := strings.SplitN(old_exec_path, "/", -1)
	exec_file_name = exec_file_name_array[len(exec_file_name_array)-1]
	new_exec_path = tmp_dir_path + "/" + exec_file_name
	// fmt.Printf("new exec path=%s\n", new_exec_path)
	/// cp cmd
	os.Link(old_exec_path, new_exec_path)

	return
}

// Create in the temp dir a Dockerfile proper to the exec type
func CreateDockerfile(tmp_dir_path string, new_exec_path string, exec_file_name string, category categorization) (imageName string) {

	imageName_arrays := strings.SplitN(exec_file_name, ".", -1)
	imageName = imageName_arrays[0]

	execPathDest := fmt.Sprintf("bin/%s", exec_file_name)
	entryPoint := fmt.Sprintf("bin/%s", exec_file_name)

	// cp the templates/Dockerfile into the tmp dir
	wd, _ := os.Getwd()
	oldTemplateDockerfilePath := wd + "/src/builder/templates/Dockerfile"
	newDockerfilePath := tmp_dir_path + "/Dockerfile"

	// read new Dockerfile
	data, err := ioutil.ReadFile(oldTemplateDockerfilePath)
	check(err)
	DockerfileString := string(data)

	// replace the placeholders to build the acutal Dockerfile
	replaceBaseImage := strings.NewReplacer("<BASE_IMAGE>", category.baseDockerImage)
	DockerfileStringReplaced := replaceBaseImage.Replace(DockerfileString)

	replaceExecPathSrc := strings.NewReplacer("<EXEC_PATH_SRC>", exec_file_name)
	DockerfileStringReplaced = replaceExecPathSrc.Replace(DockerfileStringReplaced)

	replaceExecPathDest := strings.NewReplacer("<EXEC_PATH_DEST>", execPathDest)
	DockerfileStringReplaced = replaceExecPathDest.Replace(DockerfileStringReplaced)

	replaceEntrypoint := strings.NewReplacer("<ENTRYPOINT>", entryPoint)
	DockerfileStringReplaced = replaceEntrypoint.Replace(DockerfileStringReplaced)

	// write in the Dockerfile the actual content with replaced values
	DockerfileBytesReplaced := []byte(DockerfileStringReplaced)
	err2 := ioutil.WriteFile(newDockerfilePath, DockerfileBytesReplaced, 0644)
	check(err2)

	return
}

// Launch a docker build from a Dockerfile
func DockerBuild(tmp_dir_path, imageName string) (err error) {
	fmt.Printf("Building Docker image named '%s' from Dockerfile located at %s/Dockerfile\n", imageName, tmp_dir_path)
	d, e := docker.New("tcp://192.168.59.103:2375", "")
	check(e)
	_, err = d.BuildImg(imageName, tmp_dir_path)
	check(err)
	return
}
