package builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"

	"github.com/francisbouvier/pipes/src/discovery"
	"github.com/francisbouvier/pipes/src/orch/swarm"
	"github.com/francisbouvier/pipes/src/utils"
)

const PIPES_CLIENT = "https://s3-us-west-1.amazonaws.com/pipesdocker/pipes_client"

type categorization struct {
	execType        string
	baseDockerImage string
	command         string
}

var (
	pythonCategory = categorization{
		execType:        "python",
		baseDockerImage: "python",
		command:         "python ",
	}
	rubyCategory = categorization{
		execType:        "ruby",
		baseDockerImage: "ruby",
		command:         "ruby ",
	}
	simpleBinaryCategory = categorization{
		execType:        "binary",
		baseDockerImage: "microbox/scratch",
		command:         "",
	}
)

func check(e error) {
	if e != nil {
		fmt.Printf("%s\n", e)
		panic(e)
	}
}

func BuildDockerImagesFromExec(args []string, c *cli.Context) (err error) { //execPath_category_map *map[string]categorization) {
	var exec_paths []string

	// Get input mode
	for _, arg := range args {
		arg_split_array := strings.SplitN(arg, ":", -1)
		service_path := arg_split_array[0]
		exec_paths = append(exec_paths, service_path)
		service_name_array := strings.SplitN(service_path, "/", -1)
		service_name := service_name_array[len(service_name_array)-1]
		input_mode := "stdin"
		if len(arg_split_array) > 1 {
			input_mode = arg_split_array[1]
		}
		err = WriteModeInStore(c, service_name, input_mode)
		if err != nil {
			return err
		}
	}

	execs_map, err := AssociateExecWithType(c, exec_paths)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println()

	// Iterating through the map
	for execOriginalPath, category := range execs_map {
		tmp_dir_path, new_exec_path, exec_file_name := SetTempDirectory(execOriginalPath)
		imageName := CreateDockerfile(tmp_dir_path, new_exec_path, exec_file_name, category)
		err = DockerBuild(c, tmp_dir_path, imageName)
		// Delete TempDir
		defer os.RemoveAll(tmp_dir_path)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println()
	}
	fmt.Printf("Docker images successfully built...\n")
	return
}

// Create a Dockerfile per executable passed through CLI
func AssociateExecWithType(c *cli.Context, exec_paths []string) (execPath_category_map map[string]categorization, err error) {
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
		service_name_array := strings.SplitN(exec_path, "/", -1)
		service_name := service_name_array[len(service_name_array)-1]
		command := fmt.Sprintf(execPath_category_map[exec_path].command+"%s", "/bin/"+service_name)
		err := WriteCommandInStore(c, service_name, command)
		if err != nil {
			return execPath_category_map, err
		}

	}

	return
}

func WriteCommandInStore(c *cli.Context, service_name string, command string) error {
	st, err := discovery.GetStore(c)
	if err != nil {
		return err
	}
	err = st.Write("command", command, fmt.Sprintf("services/%s", service_name))
	if err != nil {
		return err
	}
	return err
}

func WriteModeInStore(c *cli.Context, service_name string, input_mode string) error {
	st, err := discovery.GetStore(c)
	if err != nil {
		return err
	}
	err = st.Write("input_mode", input_mode, fmt.Sprintf("services/%s", service_name))
	if err != nil {
		return err
	}
	return err
}

// Set a temp directory and cp the exec in it
func SetTempDirectory(old_exec_path string) (tmp_dir_path, new_exec_path, exec_file_name string) {
	// mkdir a tmp dir
	tmp_dir_path, err := ioutil.TempDir("", "pipes_")
	check(err)

	// cp exec in that tmp dir
	info, err := os.Stat(old_exec_path)
	check(err)
	exec_file_name = info.Name()
	new_exec_path = path.Join(tmp_dir_path, exec_file_name)
	/// cp cmd
	/// TODO: check if there is no better way (buffio)
	data, err := ioutil.ReadFile(old_exec_path)
	check(err)
	err = ioutil.WriteFile(new_exec_path, data, 0755)
	check(err)

	// Add pipes_client
	p, err := utils.GetTool("pipes_client", PIPES_CLIENT)
	check(err)
	fmt.Println(p)
	data, err = ioutil.ReadFile(p)
	check(err)
	err = ioutil.WriteFile(path.Join(tmp_dir_path, "pipes_client"), data, 0755)
	check(err)

	return
}

// Create in the temp dir a Dockerfile proper to the exec type
func CreateDockerfile(tmp_dir_path string, new_exec_path string, exec_file_name string, category categorization) (imageName string) {

	imageName_arrays := strings.SplitN(exec_file_name, ".", -1)
	imageName = imageName_arrays[0]

	execPathDest := fmt.Sprintf("bin/%s", exec_file_name)
	entryPoint := fmt.Sprintf("bin/%s", exec_file_name)

	// cp the templates/Dockerfile into the tmp dir
	newDockerfilePath := tmp_dir_path + "/Dockerfile"

	// read new Dockerfile
	DockerfileString := T_GENERIC

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
func DockerBuild(c *cli.Context, tmp_dir_path, imageName string) (err error) {
	// Get store from cli
	st, err := discovery.GetStore(c)
	if err != nil {
		return err
	}

	// Get swarm from store
	sw, err := swarm.New(st)
	if err != nil {
		return err
	}

	fmt.Printf("Building Docker image named '%s' from Dockerfile located at %s/Dockerfile\n", imageName, tmp_dir_path)
	_, err = sw.BuildImg(imageName, tmp_dir_path)
	check(err)
	return
}
