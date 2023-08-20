import pytest
from pathlib import Path
from dl import samba_mount, samba_parse

class TestUtil:
    def test_samba_parse(self):
        remote_drive, local_path = samba_parse('e:/media/renders')
        assert remote_drive=='e'
        assert local_path=='/media/renders'

    def test_samba_mount(self):
        samba_root = samba_mount('bear', 'e', False)
        assert samba_root==Path('/mnt/bear/e')
        assert samba_root.is_mount()
