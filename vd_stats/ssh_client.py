import os
import paramiko


class SSHClient:
    def __init__(self, host, user, key_path, port=22, timeout=10):
        self.host = host
        self.user = user
        self.key_path = os.path.expanduser(key_path)
        self.port = port
        self.timeout = timeout
        self._client = None

    def connect(self):
        self._client = paramiko.SSHClient()
        self._client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        self._client.connect(
            self.host,
            username=self.user,
            key_filename=self.key_path,
            port=self.port,
            timeout=self.timeout,
        )

    def run(self, command):
        if not self._client:
            self.connect()
        _, stdout, stderr = self._client.exec_command(command, timeout=self.timeout)
        return stdout.read().decode().strip(), stderr.read().decode().strip()

    def close(self):
        if self._client:
            self._client.close()
            self._client = None

    def __enter__(self):
        self.connect()
        return self

    def __exit__(self, *args):
        self.close()
