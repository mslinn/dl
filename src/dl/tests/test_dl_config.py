import pytest
from dl import *

class TestDLConfig:
    def test_new(self):
        config = DLConfig(config_path='dl.config')
        assert False==True
        assert config.local
        assert config.remotes
        assert config.active_remotes
