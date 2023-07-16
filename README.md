# dl

Download videos and mp3s from many websites.
Uses `scp` to copy downloaded media to other computers that run `sshd`.

## Installation

1. Install dependencies:

   ```shell
   $ python -m pip install pyyaml
   ```

2. Copy `dl` and `dl.py` to somewhere on the path.
   An easy way to do that is to `git clone` this repository and include the directory on the `PATH`.

3. Define an environment variable called `dl` in `~/.bashrc` that points to the directory where you installed `dl.py`.

   ```
   export dl=/mnt/c/work/dl
   ```

4. Reload `~/.bashrc`:

   ```shell
   $ source ~/.bashrc
   ```

5. Copy `dl.config` to your home directory and modify to suit.
   Here is a typical `dl.config`:

   ```yaml
   local:
      mp3s: ${win_home}/Music
      vdest: ${media}/staging
      xdest: ${storage}/secret/videos

   mp3s:
      automount: /media/mslinn/Clip Jam/Music/playlist

   remotes:
      mslinn@server1:
         mp3s: /data/media/mp3s
         vdest: /data/media/staging
         xdest: /data/secret/videos
      mslinn@server2:
         disabled: true
         mp3s: /mnt/e/media/mp3s
         vdest: /mnt/e/media/staging
         xdest: /mnt/c/secret/videos
   ```

6. Verify the setup is correct by inspecting the paths in the help message:

    ```shell
    $ dl -h
    usage: dl [-h] [-d] [-v] [-x] [-V VIDEO_DEST] url

    Downloads media

    positional arguments:
      url

    options:
      -h, --help            show this help message and exit
      -d, --debug           Enable debug mode
      -v, --keep-video      Download video to /mnt/e/media/staging
      -x, --xrated          Download video to /mnt/c/secret/videos
      -V VIDEO_DEST, --video_dest VIDEO_DEST
                            download video to the specified directory

    Last modified 2023-07-16
    ```

    If you see an environment variable that was not expanded in the help message,
    that means it was not defined.
    Define the variable somehow before running this program.
