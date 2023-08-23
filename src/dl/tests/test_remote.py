import pytest
from dl import Purpose, Remote
from pathlib import Path

class TestRemote:
    def test_undefined_purpose(self):
        remote = Remote('bear')
        other = Remote('camille')

        with pytest.raises(SystemExit) as e:
            remote.compute_remote_path(other, purpose=Purpose.MP3S)
        assert e.type == SystemExit
        assert e.value.code == 'Error: Remote camille does not define a path for mp3s.'

        with pytest.raises(SystemExit) as e:
            remote.compute_remote_path(other, purpose=Purpose.VIDEOS)
        assert e.type == SystemExit
        assert e.value.code == 'Error: Remote camille does not define a path for videos.'

        with pytest.raises(SystemExit) as e:
            remote.compute_remote_path(other, purpose=Purpose.XRATED)
        assert e.type == SystemExit
        assert e.value.code == 'Error: Remote camille does not define a path for x-rated videos.'

        with pytest.raises(SystemExit) as e:
            remote.compute_remote_path(other, 'oops')
        assert e.type == SystemExit
        assert e.value.code == "Error: Unknown purpose 'oops'"

    def test_defined_purpose(self):
        remote = Remote('bear',    mp3_path=Path('/remote/mp3s'), video_path=Path('/remote/videos'), xrated_path=Path('/remote/xrated'))
        other  = Remote('camille', mp3_path=Path('/other/mp3s'),  video_path=Path('/other/videos'),  xrated_path=Path('/other/xrated'))

        assert remote.compute_remote_path(other=other, purpose=Purpose.MP3S)   == Path('/other/mp3s')
        assert remote.compute_remote_path(other=other, purpose=Purpose.VIDEOS) == Path('/other/videos')
        assert remote.compute_remote_path(other=other, purpose=Purpose.XRATED) == Path('/other/xrated')
