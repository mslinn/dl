# dl

Download videos and mp3s from many websites.

## Installation

1. Install dependencies:

   ```shell
   $ python -m pip install pyyaml
   ```

2. Copy `dl` and `dl.py` to somewhere on the path.

3. Define an environment variable called `dl` in `~/.bashrc` that points to the directory where you installed `dl.py`.
   ```
   export dl=/mnt/c/work/dl
   ```

4. Reload `~/.bashrc`:
   ```shell
   $ source ~/.bashrc
   ```

5. Copy `dl.config` to your home directory and modify to suit.

6. Verify the setup is correct:
    ```shell
    $ dl -h
    usage: dl [-h] [-d] [-v] [-x] [-V VIDEO_DEST] url

    Downloads media

    positional arguments:
      url

    options:
      -h, --help            show this help message and exit
      -d, --debug           Enable debug mode
      -v, --keep-video      Download video to ${media}/staging
      -x, --xrated          Download video to ${storage}/a_z/videos
      -V VIDEO_DEST, --video_dest VIDEO_DEST
                            download video to the specified directory

    Last modified 2023-07-16
    ```
