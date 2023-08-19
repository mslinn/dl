import pytest
from dl import *

class TestDLConfig:
    def test_new(self):
        config = DLConfig(config_path='dl.config')
        assert config.local
        assert str(config.mp3s(config.local)).endswith("/Music")
        assert str(config.vdest(config.local)).endswith("/staging")
        assert str(config.xdest(config.local)).endswith("/secret/videos")

        assert config.remotes
        assert config.active_remotes
