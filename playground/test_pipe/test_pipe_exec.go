package main
import (
    //"os"
    "fmt"
    "os/exec"
)

func pipe_commands(commands ...*exec.Cmd) ([]byte, error) {
        for i, command := range commands[:len(commands) - 1] {
            out, err := command.StdoutPipe()
            if err != nil {
                return nil, err
            }
            command.Start()
            commands[i + 1].Stdin = out
        }
        final, err := commands[len(commands) - 1].Output()
        if err != nil {
            return nil, err
        }
        return final, nil
}

func main(){

    /*
     *var dirs []string
     *if len(os.Args) > 1 {
     *    dirs = os.args[1:]
     *} else {
     *    dirs = []string{"."}
     *}
     */

    //for _, dir := range dirs {
        //c1 := exec.Command("ls", "-lrth", "var/lib/docker/containers/")
        c1 := exec.Command("ls", "-lrth")
        c2 := exec.Command("tail", "-n", "1")
        c3 := exec.Command("awk '{print $9}'")
        //output, err := pipe_commands(c1,c2,c3)

        output, err := pipe_commands(c1,c2,c3)
        if err != nil {
            fmt.Println("error")
            fmt.Errorf("%s\n",err)
        } else {
            fmt.Printf("patate")
            fmt.Printf(string(output))
        }
    //}
}

