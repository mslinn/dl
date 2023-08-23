import pytest
from dl import Remote
from pathlib import Path

class TestRemote:
    def test_undefined_purpose(self):
        remote = Remote('bear')
        other = Remote('camille')

        with pytest.raises(SystemExit) as e:
            remote.compute_remote_path(other, 'mp3s')
        assert e.type == SystemExit
        assert e.value.code == 'Remote camille does not define a path for mp3s.'

        with pytest.raises(SystemExit) as e:
            remote.compute_remote_path(other, 'vdest')
        assert e.type == SystemExit
        assert e.value.code == 'Remote camille does not define a path for videos.'

        with pytest.raises(SystemExit) as e:
            remote.compute_remote_path(other, 'xdest')
        assert e.type == SystemExit
        assert e.value.code == 'Remote camille does not define a path for x-rated videos.'

        with pytest.raises(SystemExit) as e:
            remote.compute_remote_path(other, 'oops')
        assert e.type == SystemExit
        assert e.value.code == "Unknown purpose 'oops'"

    def test_defined_purpose(self):
        remote = Remote('bear', mp3_path=Path('/remote/mp3s'), video_path=Path('/remote/videos'), x_path=Path('/remote/xrated'))
        other = Remote('camille', mp3_path=Path('/other/mp3s'), video_path=Path('/other/videos'), x_path=Path('/other/xrated'))

        assert remote.compute_remote_path(other, 'mp3s') == Path('/other/mp3s')
        assert remote.compute_remote_path(other, 'videos') == Path('/other/videos')
        assert remote.compute_remote_path(other, 'xrated') == Path('/other/xrated')
