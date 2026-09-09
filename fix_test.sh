# Fix login_passkey_test.go, login_security_test.go, and login_task_test.go conflicts by removing them from loginPage_test.go
sed -i '/func TestPasskeyLoginUsesUserBoundCeremony/,$d' handlers/auth/loginPage_test.go

# Fix user pages_test.go
git checkout -- handlers/user/pages_test.go
