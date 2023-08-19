import pytest
from dl import *

class TestDLConfig:
    def test_new(self):
        config = DLConfig(config_path='dl.config')
        assert config.local

        mp3s = str(config.mp3s(config.local))
        assert not mp3s.startswith("${win_home}")
        assert mp3s.endswith("/Music")

        vdest = str(config.vdest(config.local))
        assert not vdest.startswith("${media}")
        assert vdest.endswith("/staging")

        xdest = str(config.xdest(config.local))
        assert not xdest.startswith("${storage}")
        assert xdest.endswith("/secret/videos")

        assert config.remotes
        gojira1 = config.remotes['mslinn@gojira']
        assert gojira1['mp3s'] == '/data/media/mp3s'
        assert gojira1['vdest'] == '/data/media/staging'
        assert gojira1['xdest'] == '/data/secret/videos'

        assert config.active_remotes

        clipJam = config.active_remotes['clipJam']
        assert clipJam

        gojira2 = config.active_remotes['mslinn@gojira']
        assert gojira2
        assert gojira2['mp3s'] == '/data/media/mp3s'
        assert gojira2['vdest'] == '/data/media/staging'
        assert gojira2['xdest'] == '/data/secret/videos'

        camille = config.active_remotes['camille']
        assert camille
