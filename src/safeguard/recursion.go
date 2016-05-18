package safeguard
import(
    "os"
    "strconv"
)

var global_safeguard_level int

func GetRecursionLevel() int {
    level_text := os.Getenv("_DISTCC_SAFEGUARD")
    //log
    safeguard_level := 0
    if level_text!=""{
        safeguard_level = 1
        if tmp,err := strconv.Atoi(level_text); err != nil {
            safeguard_level = tmp
        }
    }
    //log
    global_safeguard_level = safeguard_level
    return safeguard_level
}

func IncrementRecursionLevel() error {
    new_level :=1
    if global_safeguard_level > 0 {
        new_level = global_safeguard_level+1
    }
    return os.Setenv("_DISTCC_SAFEGUARD",strconv.Itoa(new_level))
}

