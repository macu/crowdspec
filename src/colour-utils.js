// Thanks https://css-tricks.com/converting-color-spaces-in-javascript/

export const rRGB = /^rgb\((\d{1,3}), (\d{1,3}), (\d{1,3})\)$/;
export const rHSL = /^hsl\((\d{1,3}), (\d{1,3})%, (\d{1,3})%\)$/;
export const rHTML = /^#[0-9a-f]{6}$/i;

export function toRgb(input) {
	if (input && typeof input === 'object') {
		if (input.hsl) {
			return hslToRgb(input);
		}
		if (input.rgb) {
			return input;
		}
	} else if (typeof input === 'string') {
		if (rRGB.test(input)) {
			let m = rRGB.exec(input);
			return {
				r: parseInt(m[1], 10),
				g: parseInt(m[2], 10),
				b: parseInt(m[3], 10),
				rgb: true,
			};
		} else if (rHTML.test(input)) {
			function hexToInt(hex) {
				hex = hex.toLowerCase();
				let int = 0;
				for (let p = 0; p < hex.length; p++) {
					let c = hex.charCodeAt((hex.length - 1) - p);
					if (c >= 48 && c <= 57) { // 0 .. 9
						int += (c - 48) * Math.pow(16, p);
					} else if (c >= 97 && c <= 122) { // a .. z
						int += (10 + (c - 97)) * Math.pow(16, p);
					}
				}
				return int;
			}
			return {
				r: hexToInt(input.substr(1, 2)),
				g: hexToInt(input.substr(3, 2)),
				b: hexToInt(input.substr(5, 2)),
				rgb: true,
			};
		} else if (rHSL.test(input)) {
			return hslToRgb(input);
		}
	}
	return {r: 0, g: 0, b: 0, rgb: true};
}

export function toHsl(input) {
	if (input && typeof input === 'object') {
		if (input.hsl) {
			return input;
		}
		if (input.rgb) {
			return rgbToHsl(input);
		}
	} else if (typeof input === 'string') {
		if (rHSL.test(input)) {
			let m = rHSL.exec(input);
			return {
				h: parseInt(m[1], 10),
				s: parseFloat(m[2]),
				l: parseFloat(m[3]),
				hsl: true,
			};
		} else if (rRGB.test(input) || rHTML.test(input)) {
			return rgbToHsl(input);
		}
	}
	return {h: 0, s: 0, l: 0, hsl: true};
}

export function hslToRgb(hsl) {
	if (hsl && typeof hsl === 'object' && hsl.rgb) {
		return hsl;
	}

	hsl = toHsl(hsl);
	let h = hsl.h;
	let s = hsl.s / 100;
	let l = hsl.l / 100;

	let c = (1 - Math.abs(2 * l - 1)) * s;
	let x = c * (1 - Math.abs((h / 60) % 2 - 1));
	let m = l - c/2;
	let r, g, b;

	if (0 <= h && h < 60) {
		r = c; g = x; b = 0;
	} else if (60 <= h && h < 120) {
		r = x; g = c; b = 0;
	} else if (120 <= h && h < 180) {
		r = 0; g = c; b = x;
	} else if (180 <= h && h < 240) {
		r = 0; g = x; b = c;
	} else if (240 <= h && h < 300) {
		r = x; g = 0; b = c;
	} else if (300 <= h && h < 360) {
		r = c; g = 0; b = x;
	}

	return {
		r: Math.round((r + m) * 255),
		g: Math.round((g + m) * 255),
		b: Math.round((b + m) * 255),
		rgb: true,
	};
}

export function rgbToHsl(rgb) {
	if (rgb && typeof rgb === 'object' && rgb.hsl) {
		return rgb;
	}

	rgb = toRgb(rgb);
	let	r = rgb.r / 255;
	let g = rgb.g / 255;
	let b = rgb.b / 255;

	let cmin = Math.min(r, g, b);
	let cmax = Math.max(r, g, b);
	let delta = cmax - cmin;
	let h;
	if (delta == 0) {
		h = 0;
	} else if (cmax == r) {
		h = ((g - b) / delta) % 6;
	} else if (cmax == g) {
		h = (b - r) / delta + 2;
	} else {
		h = (r - g) / delta + 4;
	}

	h = Math.round(h * 60);

	let l = (cmax + cmin) / 2;

	let s = delta == 0 ? 0
		: delta / (1 - Math.abs(2*l - 1));

	return {
		h: h < 0 ? h + 360 : h,
		s: +(s * 100).toFixed(1),
		l: +(l * 100).toFixed(1),
		hsl: true,
	};
}

export function invertHsl(input) {
	let hsl = toHsl(input);

	// Stepped approach
	// Thanks https://stackoverflow.com/a/635073/1597274
	if (hsl.l <= 25) {
		hsl.l = 75; // use medium bright to contrast dark
	} else if (hsl.l <= 50) {
		hsl.l = 90; // use bright to contrast medium dark
	} else if (hsl.l <= 75) {
		hsl.l = 10; // use dark to contrast medium bright
	} else {
		hsl.l = 25; // use medium dark to contrast bright
	}

	return hsl;
}

export function encodeRgb(input) {
	let rgb = toRgb(input);
	return 'rgb(' + rgb.r + ', ' + rgb.g + ', ' + rgb.b + ')';
}

export function encodeHsl(input) {
	let hsl = toHsl(input);
	return 'hsl(' + hsl.h + ', ' + hsl.s + '%, ' + hsl.l + '%)';
}
