import os
import subprocess
import sys

os.environ['HTTP_PROXY'] = 'http://127.0.0.1:7890'
os.environ['HTTPS_PROXY'] = 'http://127.0.0.1:7890'
os.environ['NO_PROXY'] = ''

from dulwich import porcelain
from dulwich.repo import Repo

repo_path = r"D:\AI\projects\Reverse_penetration\Reverse_UsbEAmSource"
message = sys.argv[1] if len(sys.argv) > 1 else "batch commit"
remote = "https://github.com/wsolarq11/Reverse_UsbEAmSource.git"

# token
tok = subprocess.run(
    ["gh", "auth", "token"], capture_output=True, text=True, check=True
).stdout.strip()

repo = Repo(repo_path)

# stage all
porcelain.add(repo_path, paths=[b'.'])

# commit (author read from repo config)
porcelain.commit(repo_path, message=message)

# push main
porcelain.push(
    repo_path,
    remote,
    b"refs/heads/main:refs/heads/main",
    username=b"wsolarq11",
    password=tok.encode(),
    ssl_verify=False,
)
print("PUSH_OK")
