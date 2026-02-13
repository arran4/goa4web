import sys

def fix(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    new_lines = []
    i = 0
    while i < len(lines):
        line = lines[i]

        # Detect NewCoreData usage without WithSession
        if "common.NewCoreData(" in line and "WithSession" not in line:
            # Check indentation
            indent = line[:len(line) - len(line.lstrip())]

            # Check if store/session is setup in previous lines (heuristic)
            # Look back 20 lines
            has_store = False
            has_session = False
            for j in range(max(0, len(new_lines)-20), len(new_lines)):
                if "store :=" in new_lines[j] or "core.Store =" in new_lines[j]:
                    has_store = True
                if "session, _ :=" in new_lines[j] or "session :=" in new_lines[j]:
                    has_session = True

            if not has_store:
                # Insert store setup
                setup = [
                    f'{indent}store := sessions.NewCookieStore([]byte("test"))\n',
                    f'{indent}core.Store = store\n',
                    f'{indent}core.SessionName = "test-session"\n',
                    f'{indent}session, _ := store.New(req, core.SessionName)\n'
                ]
                new_lines.extend(setup)
                has_session = True # Now we have it
            elif not has_session:
                 # Store exists but no session variable?
                 # Assuming req exists.
                 new_lines.append(f'{indent}session, _ := store.New(req, core.SessionName)\n')
                 has_session = True

            # Now modify NewCoreData call
            # Assume it ends with ')' on the same line (simple case)
            if line.strip().endswith(')'):
                line = line.rstrip()[:-1] + ', common.WithSession(session))\n'
            else:
                # Multi-line call? Or comment?
                # Just append to line if it ends with )
                pass

        new_lines.append(line)
        i += 1

    with open(filepath, 'w') as f:
        f.writelines(new_lines)
    print(f"Processed {filepath}")

if __name__ == "__main__":
    fix(sys.argv[1])
