"""Build the existing PJPROJECT source and official SWIG interface in isolation (macOS)."""
from pathlib import Path
import platform
import re
import shutil
import subprocess
import sys

HERE = Path(__file__).resolve().parent
ROOT = HERE.parent
BUILD = HERE / 'build'

def run(*args):
    print('+', ' '.join(map(str, args)), flush=True)
    subprocess.run(list(map(str, args)), check=True)

if platform.system() != 'Darwin':
    raise SystemExit('This build helper is currently validated for macOS only.')
if sys.prefix == sys.base_prefix:
    raise SystemExit('Run with .venv-pjsua2/bin/python to avoid a global installation.')
cmake = shutil.which('cmake') or '/opt/homebrew/bin/cmake'
swig = Path(sys.executable).parent / 'swig'
if not swig.exists():
    raise SystemExit('Install swig in the virtual environment first.')
options = []
cache = ROOT / 'build' / 'CMakeCache.txt'
if cache.exists():
    for line in cache.read_text().splitlines():
        match = re.fullmatch(r'((?:PJ[A-Z0-9_]*|SRTP_[A-Z0-9_]*)):(BOOL|STRING)=(.*)', line)
        if match:
            options.append('-D' + match[1] + ':' + match[2] + '=' + match[3])
run(cmake, '-S', HERE, '-B', BUILD, '-DCMAKE_BUILD_TYPE=Release',
    '-DPython3_EXECUTABLE=' + sys.executable,
    '-DSWIG_EXECUTABLE=' + str(swig), *options)
run(cmake, '--build', BUILD, '--target', '_pjsua2', '--parallel', '4')
run(cmake, '--install', BUILD, '--component', 'PythonBindings')
run(sys.executable, HERE / 'smoke_test.py')
