import pytest
from dl import *

class TestMediaFile:
    def test_media_file(self):
      mf = MediaFile(config='dl.config', path=Path('/blah/ick.mp4'))
      assert mf.is_mp3==False
      assert mf.is_video==True
      assert mf.file_type=='mp4'
