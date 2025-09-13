import os
import shutil
import subprocess
import sys
from pathlib import Path

DIST_DIR = Path(__file__).parent / "dist"
SERVER_STATIC_DIR = Path(__file__).parent.parent / "server" / "static"


def run_command(cmd, cwd=None):
    """Run a shell command and exit if it fails."""
    print(f"Running: {' '.join(cmd)}")
    result = subprocess.run(" ".join(cmd), cwd=cwd, shell=True)
    if result.returncode != 0:
        print(f"Command failed: {' '.join(cmd)}", file=sys.stderr)
        sys.exit(result.returncode)


def main():
    script_dir = Path(__file__).parent.resolve()
    os.chdir(script_dir)
    run_command(["npx", "tsc", "-b"])
    run_command(["npx", "vite", "build"])
    if SERVER_STATIC_DIR.exists():
        shutil.rmtree(SERVER_STATIC_DIR)
    shutil.copytree(DIST_DIR, SERVER_STATIC_DIR)
    print("Build complete.")


if __name__ == "__main__":
    main()
