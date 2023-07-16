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
    dl - Download videos from various sites.
    Can save video, or can just save audio as mp3.

    Usage: dl [options] url

    Options are:
      -h       display this message and quit
      -v       download video to /home/mslinn/Videos
      -V dir   download video to any given directory
      -x       download video to /home/mslinn/Videos

    Last modified 2022-07-24
    ```
