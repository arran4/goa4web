sed -i 's/for _, id := range rows {/ids = append(ids, rows...)/g' handlers/admin/retry_sent_task.go
sed -i '/ids = append(ids, id)/d' handlers/admin/retry_sent_task.go
