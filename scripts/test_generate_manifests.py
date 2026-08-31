import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


class ManifestGeneratorTest(unittest.TestCase):
    def test_generates_all_ecosystems_without_placeholders(self):
        root = Path(__file__).resolve().parents[1]
        names = [
            "switchboard_windows_amd64.zip",
            "switchboard_windows_arm64.zip",
            "switchboard_darwin_amd64.tar.gz",
            "switchboard_darwin_arm64.tar.gz",
            "switchboard_linux_amd64.tar.gz",
            "switchboard_linux_arm64.tar.gz",
            "switchboard_linux_armv7.tar.gz",
        ]
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            checksums = directory / "checksums.txt"
            checksums.write_text("".join(f"{'a' * 64}  {name}\n" for name in names), encoding="utf-8")
            output = directory / "distribution"
            subprocess.run(
                [sys.executable, str(root / "scripts" / "generate_manifests.py"), "--version", "v1.2.3", "--checksums", str(checksums), "--output", str(output)],
                check=True,
            )
            files = list(output.rglob("*"))
            self.assertTrue(any(path.name == "switchboard.rb" for path in files))
            self.assertTrue(any(path.name == "switchboard.json" for path in files))
            self.assertTrue(any(path.name == "PKGBUILD" for path in files))
            for path in files:
                if path.is_file():
                    self.assertNotIn("{{", path.read_text(encoding="utf-8"))


if __name__ == "__main__":
    unittest.main()
