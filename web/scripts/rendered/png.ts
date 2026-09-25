/**
 * A PNG reader for the one shape Chromium's screenshots take: 8-bit
 * truecolor, with or without alpha, non-interlaced. Any other shape is
 * refused rather than guessed at, so a sampled pixel is always the
 * byte the browser wrote.
 */
import { inflateSync } from 'node:zlib';

const SIGNATURE = [0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a];

export interface Image {
  width: number;
  height: number;
  channels: number;
  /** Unfiltered scanlines, `width * channels` bytes each. */
  pixels: Uint8Array;
}

export function decodePNG(png: Uint8Array): Image {
  if (png.length < SIGNATURE.length || SIGNATURE.some((b, i) => png[i] !== b)) {
    throw new Error('png: not a PNG file');
  }
  const view = new DataView(png.buffer, png.byteOffset, png.byteLength);
  let width = 0;
  let height = 0;
  let channels = 0;
  const idat: Uint8Array[] = [];
  let pos = SIGNATURE.length;
  while (pos + 8 <= png.length) {
    const length = view.getUint32(pos);
    const type = String.fromCharCode(...png.subarray(pos + 4, pos + 8));
    const start = pos + 8;
    const data = png.subarray(start, start + length);
    pos = start + length + 4;
    if (type === 'IHDR') {
      width = view.getUint32(start);
      height = view.getUint32(start + 4);
      const [bitDepth, colorType, , , interlace] = data.subarray(8, 13);
      if (bitDepth !== 8 || (colorType !== 2 && colorType !== 6) || interlace !== 0) {
        throw new Error(
          `png: bit depth ${bitDepth}, color type ${colorType}, interlace ${interlace}; only 8-bit non-interlaced truecolor is read`
        );
      }
      channels = colorType === 6 ? 4 : 3;
    } else if (type === 'IDAT') {
      idat.push(data);
    } else if (type === 'IEND') {
      break;
    }
  }
  if (channels === 0) throw new Error('png: no IHDR chunk');
  const raw = inflateSync(Buffer.concat(idat));
  const stride = width * channels;
  if (raw.length !== height * (stride + 1)) {
    throw new Error(`png: ${raw.length} inflated bytes, want ${height * (stride + 1)}`);
  }
  const pixels = new Uint8Array(height * stride);
  for (let y = 0; y < height; y++) {
    const filter = raw[y * (stride + 1)];
    const line = raw.subarray(y * (stride + 1) + 1, (y + 1) * (stride + 1));
    const out = pixels.subarray(y * stride, (y + 1) * stride);
    const prev = y > 0 ? pixels.subarray((y - 1) * stride, y * stride) : new Uint8Array(stride);
    for (let i = 0; i < stride; i++) {
      const left = i >= channels ? out[i - channels] : 0;
      const up = prev[i];
      const upLeft = i >= channels ? prev[i - channels] : 0;
      out[i] = (line[i] + predictor(filter, left, up, upLeft)) & 0xff;
    }
  }
  return { width, height, channels, pixels };
}

// PNG section 9.2's five filter types.
function predictor(filter: number, left: number, up: number, upLeft: number): number {
  switch (filter) {
    case 0:
      return 0;
    case 1:
      return left;
    case 2:
      return up;
    case 3:
      return (left + up) >> 1;
    case 4: {
      const p = left + up - upLeft;
      const pa = Math.abs(p - left);
      const pb = Math.abs(p - up);
      const pc = Math.abs(p - upLeft);
      if (pa <= pb && pa <= pc) return left;
      return pb <= pc ? up : upLeft;
    }
    default:
      throw new Error(`png: unknown filter type ${filter}`);
  }
}

/** The pixel at (`x`, `y`) as lowercase `#rrggbb`; alpha is dropped. */
export function pixelHex(image: Image, x: number, y: number): string {
  if (x < 0 || y < 0 || x >= image.width || y >= image.height) {
    throw new Error(`png: (${x}, ${y}) is outside a ${image.width}x${image.height} image`);
  }
  const at = (y * image.width + x) * image.channels;
  const rgb = image.pixels.subarray(at, at + 3);
  return `#${[...rgb].map((b) => b.toString(16).padStart(2, '0')).join('')}`;
}
