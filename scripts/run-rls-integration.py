import os
import getpass
import shutil
import socket
import subprocess
import tempfile
from pathlib import Path

REPO = Path(__file__).resolve().parents[1]

def run(args, **kwargs):
    subprocess.run(args, check=True, **kwargs)

with tempfile.TemporaryDirectory(prefix='zenres-rls-test-') as raw:
    temp = Path(raw)
    data = temp / 'data'
    socket_dir = temp / 'socket'
    socket_dir.mkdir()
    with socket.socket() as sock:
        sock.bind(('127.0.0.1', 0))
        port = sock.getsockname()[1]
    run([shutil.which('initdb'), '-D', str(data), '--auth=trust', '--no-instructions'], stdout=subprocess.DEVNULL)
    started = False
    try:
        run([shutil.which('pg_ctl'), '-D', str(data), '-o', f'-p {port} -k {socket_dir} -c listen_addresses=', '-w', 'start'], stdout=subprocess.DEVNULL)
        started = True
        host = ['-h', str(socket_dir), '-p', str(port)]
        run([shutil.which('createdb'), *host, 'zenres_rls_test'])
        for name in ['001_auth_and_pgp_vault.sql', '002_zevra_core_platform.sql']:
            run([shutil.which('psql'), *host, '-d', 'zenres_rls_test', '-v', 'ON_ERROR_STOP=1', '-f', str(REPO / 'db' / 'migrations' / name)], stdout=subprocess.DEVNULL)
        env = os.environ.copy()
        env['RLS_TEST_DATABASE_URL'] = f'host={socket_dir} port={port} user={getpass.getuser()} dbname=zenres_rls_test sslmode=disable'
        run(['go', 'test', './core-api/internal/rls', '-run', 'TestTenantPolicyIsolation', '-count=1', '-v'], cwd=REPO / 'services', env=env)
    finally:
        if started:
            run([shutil.which('pg_ctl'), '-D', str(data), '-m', 'immediate', '-w', 'stop'], stdout=subprocess.DEVNULL)
