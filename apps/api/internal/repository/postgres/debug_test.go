//go:build ignore

package postgres

import (
    "fmt"
)

func DebugListTasks() {
    repo, _ := New("postgresql://postgres:pass@172.20.0.16:5432/auditer?sslmode=disable")
    defer repo.Close()
    
    tasks := repo.ListTasks("", "")
    fmt.Printf("ListTasks returned: %d\n", len(tasks))
    for i, t := range tasks {
        if i >= 5 {
            fmt.Printf("... and %d more\n", len(tasks)-5)
            break
        }
        fmt.Printf("  [%d] %s | %s\n", i+1, t.ID, t.Name)
    }
}
