package main

import (
    "os/exec"
    "errors"
    "strings"
    "os"
    "fmt"
)
//something is wrong with the way I handle quotes and escapes
var builtins = []string{"type", "echo", "exit", "pwd"}

func ParseCommand(command string) (string,[]string,string,string,int){
    var (
	args []string
	tmp string
	isEsc bool
	inSQuote bool
	inDQuote bool
	isHome bool
	Mode int
	getFile int = -1
	OutFile string = ""
	ErrFile string = ""
    )
    for i:= 0 ;i < len(command);i++{
	c := command[i]
	if isHome {
	    if c == ' ' || c == '/' {
		tmp += os.Getenv("HOME")
	    } else {
		tmp += "~"
	    }
	    isHome = false
	}
	if isEsc {
	    if inDQuote && !(c == '$' || c == '\\' || c == '"'){
		tmp += "\\"
	    }
	    tmp = tmp + string(c)
	    isEsc = false
	} else if c == '>' && !inSQuote && !inDQuote {
	    j := len(tmp) -1
	    if j >= 0 && tmp[j] == '>'{
		tmp = ">>"
	    } else {
		if j >= 0 && tmp == "2" {
		    getFile = 1
		} else {
		    getFile = 0
		}
		if tmp != "2" && tmp != "1" && tmp != ""{
		    args = append(args,tmp)
		}
		tmp = ">"
	    }
	} else if tmp == ">" || tmp == ">>" {
	    if tmp == ">"{
		if getFile == 1 {
		    Mode = Mode &^ 1
		}else {
		    Mode = Mode &^ 2
		}
	    }else {
		if getFile == 1 {
		    Mode |= 1
		}else {
		    Mode |= 2
		}
	    }
	    tmp = ""
	    i--
	} else if c == '~' && !inDQuote && !inSQuote && tmp == ""{
	    isHome = true
	} else if c == '\\' && !inSQuote {
	    isEsc = true
	} else if c == '\'' && !inDQuote {
	    inSQuote = !inSQuote
	} else if c == '"' && !inSQuote {
	    inDQuote = !inDQuote
	} else if c == ' ' && !inSQuote && !inDQuote{
	    if tmp != "" {
		if getFile != -1 {
		    if getFile == 0{
			OutFile = tmp
		    } else {
			ErrFile = tmp
		    }
		    getFile = -1
		} else {
		    args = append(args,tmp)
		}	
		tmp = ""
	    }
	} else {
	    tmp = tmp + string(c)
	}
    }

    if tmp != "" || isHome{
	if isHome {
	    tmp = os.Getenv("HOME")
	}
	if getFile != -1{
	    if getFile == 0 {
		OutFile = tmp
	    } else {
		ErrFile = tmp
	    }
	    getFile = -1
	}else {
	    args = append(args,tmp)
	}
    }

    if len(args) == 0 {
	return "",[]string{},OutFile,ErrFile,Mode
    }
    return args[0],args[1:],OutFile,ErrFile,Mode
}


func ChangeDirectory(path string) {
    err := os.Chdir(path)
    if err != nil{
	fmt.Println("cd: " +  path + ": No such file or directory")
    }
}

func GetUserCommand() (string, error){
    fmt.Fprint(os.Stdout, "$ ")
    command, err := ReadInput()
    if err != nil {
	return "", err
    }
    return command,nil
}

func Execute(cmd string, args []string,outf string,errf string,mods int) error {
    var outfile,errfile *os.File
    for _,path := range (strings.Split(os.Getenv("PATH"),":")){
	file := path + "/" + cmd
	if _,err := os.Stat(file); err == nil {
	    command := exec.Command(cmd,args...)

	    if outf == ""{
		outfile = os.Stdout
	    } else {
		if mods & 2 == 0{
		    outfile,_ = os.OpenFile(outf,os.O_CREATE | os.O_TRUNC | os.O_WRONLY, 0644)
		}else{
		    outfile,_ = os.OpenFile(outf,os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0644)
		}
	    }
	    if errf == ""{
		errfile = os.Stdout
	    } else {
		if mods & 1 == 0{
		    errfile,_ = os.OpenFile(errf,os.O_CREATE | os.O_TRUNC | os.O_WRONLY, 0644)
		}else{
		    errfile,_ = os.OpenFile(errf,os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0644)
		}
	    }
	    command.Stdout = outfile
	    command.Stderr = errfile
	    command.Run()
	    return nil
	}
    }
    return errors.New(fmt.Sprintf("%s: command not found\n",cmd))
}

func WriteToTarget(output string,file string, mod int) {
    var f *os.File
    if file == "" {
	os.Stdout.Write([]byte(output))
	return
    }
    if mod != 0{
	f, _ = os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
    }else {
    f, _ = os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
    }
    f.Write([]byte(output))
    f.Close()
}
func FindCmd(cmd string) string {
    for _,builtin := range builtins {
	if builtin == cmd {
	    return fmt.Sprintf("%s is a shell builtin",cmd)
	}
    }
    for _,path := range (strings.Split(os.Getenv("PATH"), ":")) {
	file := path + "/" + cmd
	if _, err := os.Stat(file); err == nil{
	    return fmt.Sprintf("%s is %s",cmd,file)
	}
    }
    return fmt.Sprintf("%s: not found",cmd)
}
