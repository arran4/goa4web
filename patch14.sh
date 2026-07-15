sed -i 's/for _, id := range rows {/ids = append(ids, rows...)/g' handlers/admin/resend_queue_task.go
sed -i '/ids = append(ids, id)/d' handlers/admin/resend_queue_task.go
