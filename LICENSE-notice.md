# RethinkRAW - Commercial use

RethinkRAW is provided **free of charge** for **personal**, **non-commercial** use.

See [LICENSE.md](LICENSE.md).

Additionally, the source code to RethinkRAW is made available for all to inspect and learn from.
But, because of restrictions on commercial use,
RethinkRAW is **not** “free and open-source software.”

The rationale behind this decision is as follows.

RethinkRAW relies heavily on other software to do what it does.

**In particular**, it relies on:
<dl>
  <dt>Adobe DNG Converter</dt>
  <dd>to convert RAW photos to DNG and render previews of those</dd>
  <dt>dcraw by Dave Coffin</dt>
  <dd>to extract RAW data and previews from RAW photos and DNGs</dd>
  <dt>ExifTool by Phil Harvey</dt>
  <dd>to read/write/edit metadata in RAW photos, DNGs and JPEGs</dd>
</dl>

**ExifTool** and **dcraw** are both free and open source software.
Using them commercially is allowed,
but you should consider donating to those projects if you do so:
- https://exiftool.org/#donate
- https://www.dechifro.org/dcraw/

Adobe DNG Converter, which is free as in _gratis_, but not _libre_,
can (and **must**) be downloaded from:<br>
https://helpx.adobe.com/photoshop/using/adobe-dng-converter.html

Adobe invests heavily in its photography software.
DNG Converter decodes RAW photos from hundreds of cameras
(with calibration profiles for hundreds of cameras and lenses),
and generates high quality, full size previews of them.

If you’re making money from photography, and find RethinkRAW useful,
you should license _some_ Adobe photography software yourself.
If you hold a valid license to any of Photoshop/Lightroom/Elements,
you’re free to use my comparatively small contribution however you want,
_including commercially_.