sed -i 's/for _, id := range rows {\n\t\t\t\t\tids = append(ids, id)\n\t\t\t\t}/ids = append(ids, rows...)/g' handlers/admin/resend_queue_task.go
