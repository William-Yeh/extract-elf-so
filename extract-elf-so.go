// Extract .so files from specified ELF executables, and pack them in a tarball.
//
//
// Copyright 2015 William Yeh <william.pjyeh@gmail.com>. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"fmt"

	"bytes"
	"regexp"
	"strings"

	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/docopt/docopt-go"
)

var REGEX_VDSO = regexp.MustCompile(`linux-vdso.so`)
var REGEX_SO_FILE = regexp.MustCompile(`\s(\/[^\s]+)\s+\([^)]+\)$`)

var NSS_NET_SO_FILES = []string{
	"/lib/x86_64-linux-gnu/libresolv.so.2",
	"/usr/lib/libdns.so.100",
	"/lib/x86_64-linux-gnu/libnss_dns.so.2",
	"/lib/x86_64-linux-gnu/libnss_files.so.2",
	"/lib/x86_64-linux-gnu/libnss_myhostname.so.2"}

var NSS_NET_ETC_FILES = []string{
	"/etc/nsswitch.conf",
	"/etc/services"}

var TARBALL_FILENAME string = ""
var TAR_COMPRESSION_MODE string = "-cvf"

const USAGE string = `Extract .so files from specified ELF executables, and pack them in a tarball.

Usage:
  extract-elf-so  [options]  [(--add <so_file>)...]  <elf_files>...
  extract-elf-so  --help
  extract-elf-so  --version

Options:
  -d <dir>, --dest <dir>          Destination dir in the tarball to place the elf_files;
                                    [default: /usr/local/bin].
  -n <name>, --name <name>        Generated tarball filename (without .gz/.tar.gz);
                                    [default: rootfs].
  -a <so_file>, --add <so_file>   Additional .so files to be appended into the tarball.
  -s <so_dir>, --sodir <so_dir>   Directory in the tarball to place additional .so files;
                                    [default: /usr/lib].
  -z                              Compress the output tarball using gzip.
  --nss-net                       Install networking stuff of NSS;  [default: false].

`

func main() {
	arguments := process_cmdline()

	ldd_output := collect_ldd_output(arguments["<elf_files>"].([]string))

	so_filelist := extract_so_files(ldd_output)
	//fmt.Println(so_filelist)

	output_files(arguments, so_filelist)
}

// This func parses and validates cmdline args
func process_cmdline() map[string]interface{} {

	arguments, _ := docopt.Parse(USAGE, nil, true, "0.1", false)

	// validate elf_files
	for _, filename := range arguments["<elf_files>"].([]string) {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			fmt.Printf("Error: no such file or directory: %s", filename)
			os.Exit(1)
		}
	}

	// handle tarball name
	if arguments["-z"].(bool) {
		TARBALL_FILENAME = arguments["--name"].(string) + ".tar.gz"
		TAR_COMPRESSION_MODE = "-zcvf"
	} else {
		TARBALL_FILENAME = arguments["--name"].(string) + ".tar"
	}

	return arguments
}

func collect_ldd_output(elf_files []string) string {
	var buffer bytes.Buffer

	for _, filename := range elf_files {
		out, err := exec.Command("ldd", filename).Output()
		if err != nil {
			fmt.Printf("Error for %s - %s", filename, err)
			os.Exit(1)
		}
		buffer.Write(out) // append to buffer
	}

	return buffer.String()
}

func extract_so_files(ldd_output string) []string {
	var filelist = make([]string, 50)

	for _, line := range strings.Split(ldd_output, "\n") { // for each line
		// ignore vDSO files
		if result := REGEX_VDSO.FindStringSubmatch(line); result != nil {
			continue
		}

		//fmt.Println("---> ", line)
		if result := REGEX_SO_FILE.FindStringSubmatch(line); result != nil {
			filelist = append(filelist, result[1])
		}
	}

	RemoveDuplicates(&filelist)
	return filelist
}

func output_files(arguments map[string]interface{}, so_filelist []string) {

	var tarball_filelist = make([]string, 50)

	//
	// create temp output dir
	//
	temp_dir, err := ioutil.TempDir("", "extractelfso")
	if err != nil {
		checkError(err)
	}
	defer os.RemoveAll(temp_dir)

	//
	// copy ELF file(s) to temp output directory...
	//
	exe_dest_dir := path.Join(temp_dir, arguments["--dest"].(string))
	if err := os.MkdirAll(exe_dest_dir, 0755); err != nil {
		checkError(err)
	}
	for _, file := range arguments["<elf_files>"].([]string) {
		file_basename := path.Base(file)
		dest_file_relpath := path.Join(arguments["--dest"].(string), file_basename)
		dest_file_fullpath := path.Join(exe_dest_dir, file_basename)

		tarball_filelist = append(tarball_filelist, dest_file_relpath[1:]) // remove heading '/' char

		if exec.Command("cp", "-rf", "--dereference", file, exe_dest_dir).Run() != nil {
			checkError(err)
		}
		if os.Chmod(dest_file_fullpath, 0755) != nil {
			checkError(err)
		}
	}
	//fmt.Println(tarball_filelist)

	//
	// append --nss-net files, if necessary...
	//
	if arguments["--nss-net"].(bool) {
		tarball_filelist = append(tarball_filelist, NSS_NET_SO_FILES...)
		tarball_filelist = append(tarball_filelist, NSS_NET_ETC_FILES...)
	}
	//fmt.Println(tarball_filelist)

	//
	// append .so files deduced from ELF file(s)...
	//
	tarball_filelist = append(tarball_filelist, so_filelist...)
	//fmt.Println(tarball_filelist)

	//
	// copy additional .so file(s)...
	//
	if len(arguments["--add"].([]string)) > 0 {
		so_dest_dir := path.Join(temp_dir, arguments["--sodir"].(string))
		if err := os.MkdirAll(so_dest_dir, 0755); err != nil {
			checkError(err)
		}

		for _, file := range arguments["--add"].([]string) {
			if exec.Command("cp", "-rf", "--dereference", file, so_dest_dir).Run() != nil {
				checkError(err)
			}

			dest_file_basename := path.Base(file)
			dest_file_relpath := path.Join(arguments["--sodir"].(string), dest_file_basename)
			tarball_filelist = append(tarball_filelist, dest_file_relpath[1:]) // remove heading '/' char
		}
	}

	//
	// generate tarball...
	//
	pwd, _ := os.Getwd()
	rootfs_tarball_fullpath := path.Join(pwd, TARBALL_FILENAME)
	RemoveDuplicates(&tarball_filelist)
	output_tarball(rootfs_tarball_fullpath, compactArray(tarball_filelist), temp_dir)
}

func output_tarball(tarball_fullpath string, tarball_filelist []string, working_dir string) {

	dumpArray(tarball_filelist)

	pwd, _ := os.Getwd()
	if err := os.Chdir(working_dir); err != nil {
		checkError(err)
	}
	defer os.Chdir(pwd)

	cmd_args := []string{"--dereference", TAR_COMPRESSION_MODE, tarball_fullpath}
	cmd_args = append(cmd_args, tarball_filelist...)
	fmt.Println(cmd_args)

	if err := exec.Command("tar", cmd_args...).Run(); err != nil {
		checkError(err)
	}
}

// remove duplicates in a slice.
// @see https://groups.google.com/d/msg/golang-nuts/-pqkICuokio/ZfSRfU_CdmkJ
func RemoveDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

// compact nil slice items
func compactArray(entries []string) []string {
	new_entries := entries[:0]
	for _, x := range entries {
		if len(x) > 0 {
			new_entries = append(new_entries, x)
		}
	}
	return new_entries
}

func dumpArray(data []string) {
	for _, item := range data {
		fmt.Println("'" + item + "'")
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
