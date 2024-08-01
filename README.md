# VOICEVOX to SRT

## Usage

1. At first, export wav + txt sequential files from VOICEVOX
2. Run this command on the exported directory

```sh
voicevox2srt
```

3. By default, `subtitle.srt` should be generated in the same directory

### Options

Specify output file name:

```sh
voicevox2srt -o sample.srt
```

Specify target directory:

```sh
voicevox2srt target/dir
```