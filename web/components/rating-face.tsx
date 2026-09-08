"use client";

import { useId } from "react";
import type { ReactNode } from "react";

const art = (name: string) => `var(--hv-art-${name})`;

const faceColors = {
  1: [art("ff7a68"), art("e8456b")],
  2: [art("ffb45c"), art("ff7859")],
  3: [art("ffd66b"), art("f3a94b")],
  4: [art("7ed6b2"), art("2fa58f")],
  5: [art("5d918a"), art("315759")]
} as const;

export const RATING_FACE_SETS = [
  { id: "heard", value: "heard", label: "heard signature", description: "Our warm, colorful default" },
  { id: "clay", value: "clay", label: "Soft clay", description: "Rounded, dimensional, and friendly" },
  { id: "glass", value: "glass", label: "Glossy glass", description: "Translucent gradients with a soft glow" },
  { id: "minimal", value: "minimal", label: "Minimal line", description: "Clean teal and gold line art" },
  { id: "retro", value: "retro", label: "Retro sticker", description: "Bold, playful, die-cut character art" },
  { id: "iphone", value: "iphone", label: "iPhone emoji", description: "Classic iOS-style yellow emoji faces" }
] as const;

export type RatingValue = keyof typeof faceColors;
export type RatingFaceSet = (typeof RATING_FACE_SETS)[number]["value"];

type RatingFaceProps = {
  className?: string;
  faceSet?: RatingFaceSet;
  rating: RatingValue;
};

export function RatingFace({ className = "", faceSet = "heard", rating }: RatingFaceProps) {
  const id = useId().replaceAll(":", "");
  if (faceSet === "clay") return <ClayFace className={className} id={id} rating={rating} />;
  if (faceSet === "glass") return <GlassFace className={className} id={id} rating={rating} />;
  if (faceSet === "minimal") return <MinimalFace className={className} rating={rating} />;
  if (faceSet === "retro") return <RetroFace className={className} id={id} rating={rating} />;
  if (faceSet === "iphone") return <IphoneAssetFace className={className} rating={rating} />;
  return <HeardFace className={className} id={id} rating={rating} />;
}

function SvgFrame({ children, className, label }: { children: ReactNode; className: string; label: string }) {
  return (
    <svg aria-label={label} className={className} focusable="false" role="img" viewBox="0 0 100 106" xmlns="http://www.w3.org/2000/svg">
      {children}
    </svg>
  );
}

function HeardFace({ className, id, rating }: { className: string; id: string; rating: RatingValue }) {
  const [start, end] = faceColors[rating];
  return (
    <SvgFrame className={className} label={`heard signature rating ${rating}`}>
      <defs><linearGradient id={`heard-${id}`} x1="18" x2="82" y1="12" y2="92" gradientUnits="userSpaceOnUse"><stop stopColor={start}/><stop offset="1" stopColor={end}/></linearGradient></defs>
      <ellipse cx="50" cy="96" fill={art("19383a")} opacity="0.12" rx="31" ry="6"/>
      <circle cx="50" cy="50" fill={`url(#heard-${id})`} r="43" stroke={art("ffffff")} strokeWidth="3"/>
      <path d="M24 31C31 18 46 12 60 16" fill="none" opacity="0.28" stroke={art("ffffff")} strokeLinecap="round" strokeWidth="7"/>
      <HeardExpression rating={rating}/>
    </SvgFrame>
  );
}

function HeardExpression({ rating }: { rating: RatingValue }) {
  if (rating === 1) return <SoftSadExpression ink={art("19383a")}/>;
  if (rating === 2) return <g fill="none" stroke={art("19383a")} strokeLinecap="round"><path d="M31 47q6-5 12 0M57 47q6-5 12 0" strokeWidth="4"/><path d="M35 73Q50 61 65 73" strokeWidth="5"/></g>;
  if (rating === 3) return <g fill={art("19383a")} stroke={art("19383a")} strokeLinecap="round"><circle cx="37" cy="48" r="4" stroke="none"/><circle cx="63" cy="48" r="4" stroke="none"/><path d="M37 70h26" fill="none" strokeWidth="5"/></g>;
  if (rating === 4) return <g fill="none" stroke={art("19383a")} strokeLinecap="round"><path d="M31 48q6 6 12 0M57 48q6 6 12 0" strokeWidth="4"/><path d="M33 65q17 20 34 0" fill={art("fff7e8")} strokeLinejoin="round" strokeWidth="4"/></g>;
  return <g strokeLinecap="round" strokeLinejoin="round"><path d="M38 58C35.8 56.1 28 50.7 28 44.8c0-3.8 2.8-6.8 6.6-6.8 2.3 0 4.3 1.1 5.4 2.9 1.1-1.8 3.1-2.9 5.4-2.9 3.8 0 6.6 3 6.6 6.8 0 5.9-7.8 11.3-10 13.2l-2 1.8Z" fill={art("fff3d7")} stroke={art("19383a")} strokeWidth="2.5"/><path d="M60 58C57.8 56.1 50 50.7 50 44.8c0-3.8 2.8-6.8 6.6-6.8 2.3 0 4.3 1.1 5.4 2.9 1.1-1.8 3.1-2.9 5.4-2.9 3.8 0 6.6 3 6.6 6.8 0 5.9-7.8 11.3-10 13.2l-2 1.8Z" fill={art("fff3d7")} stroke={art("19383a")} strokeWidth="2.5"/><path d="M31 66q19 21 38 0" fill={art("fff8ec")} stroke={art("19383a")} strokeWidth="4"/></g>;
}

const clayColors: Record<RatingValue, [string, string]> = {
  1: [art("e97774"), art("b94555")], 2: [art("efa071"), art("d36d61")], 3: [art("efc56d"), art("ce914d")], 4: [art("84c8a2"), art("438d78")], 5: [art("f1aaa4"), art("cc6f7a")]
};

function ClayFace({ className, id, rating }: { className: string; id: string; rating: RatingValue }) {
  const [light, dark] = clayColors[rating];
  return (
    <SvgFrame className={className} label={`soft clay rating ${rating}`}>
      <defs><radialGradient id={`clay-${id}`} cx="34%" cy="25%" r="76%"><stop stopColor={art("fff5df")}/><stop offset=".3" stopColor={light}/><stop offset="1" stopColor={dark}/></radialGradient><filter id={`clay-shadow-${id}`} x="-30%" y="-30%" width="160%" height="170%"><feDropShadow dx="0" dy="5" floodColor={art("5b2733")} floodOpacity=".24" stdDeviation="4"/></filter></defs>
      <ellipse cx="50" cy="97" fill={art("512a31")} opacity=".16" rx="29" ry="5"/>
      <circle cx="50" cy="49" fill={`url(#clay-${id})`} filter={`url(#clay-shadow-${id})`} r="40"/>
      <ellipse cx="35" cy="63" fill={art("ffb4aa")} opacity=".62" rx="7" ry="4"/><ellipse cx="65" cy="63" fill={art("ffb4aa")} opacity=".62" rx="7" ry="4"/>
      <ClayExpression rating={rating}/><ellipse cx="37" cy="27" fill={art("fff")} opacity=".35" rx="12" ry="6" transform="rotate(-28 37 27)"/>
    </SvgFrame>
  );
}

function ClayExpression({ rating }: { rating: RatingValue }) {
  const ink = art("552c36");
  if (rating === 1) return <SoftSadExpression ink={ink}/>;
  if (rating === 2) return <g fill="none" stroke={ink} strokeLinecap="round"><ellipse cx="37" cy="49" rx="3.5" ry="5" fill={ink} stroke="none"/><ellipse cx="63" cy="49" rx="3.5" ry="5" fill={ink} stroke="none"/><path d="M35 73q15-11 30 0" strokeWidth="5"/></g>;
  if (rating === 3) return <g fill={ink} stroke={ink} strokeLinecap="round"><circle cx="37" cy="49" r="4"/><circle cx="63" cy="49" r="4"/><path d="M38 71h24" fill="none" strokeWidth="5"/></g>;
  if (rating === 4) return <g fill="none" stroke={ink} strokeLinecap="round"><path d="M31 48q6 7 12 0m14 0q6 7 12 0" strokeWidth="4.5"/><path d="M34 65q16 17 32 0" fill={art("fff1dc")} strokeWidth="4"/></g>;
  return <g><Heart cx={37} cy={48} color={art("552c36")} fill={art("cf4e63")}/><Heart cx={63} cy={48} color={art("552c36")} fill={art("cf4e63")}/><path d="M31 65q19 20 38 0" fill={art("fff4df")} stroke={ink} strokeLinecap="round" strokeWidth="4"/></g>;
}

function GlassFace({ className, id, rating }: { className: string; id: string; rating: RatingValue }) {
  const hue = rating <= 2 ? [art("ff8bb5"), art("8d6cff")] : rating === 3 ? [art("74d7ef"), art("8f7cff")] : [art("58e2e0"), art("906dff")];
  return (
    <SvgFrame className={className} label={`glossy glass rating ${rating}`}>
      <defs><radialGradient id={`glass-${id}`} cx="28%" cy="20%" r="90%"><stop stopColor={art("ffffff")} stopOpacity=".94"/><stop offset=".26" stopColor={hue[0]} stopOpacity=".76"/><stop offset=".72" stopColor={hue[1]} stopOpacity=".82"/><stop offset="1" stopColor={art("263a90")} stopOpacity=".9"/></radialGradient><filter id={`glass-glow-${id}`} x="-45%" y="-45%" width="190%" height="200%"><feDropShadow dx="0" dy="5" floodColor={hue[1]} floodOpacity=".38" stdDeviation="6"/></filter></defs>
      <ellipse cx="50" cy="97" fill={hue[1]} opacity=".18" rx="31" ry="5"/><circle cx="50" cy="49" fill={`url(#glass-${id})`} filter={`url(#glass-glow-${id})`} r="41" stroke={art("fff")} strokeOpacity=".7" strokeWidth="1.5"/><path d="M24 31c8-15 25-19 39-12" fill="none" stroke={art("fff")} strokeLinecap="round" strokeOpacity=".58" strokeWidth="7"/><GlassExpression rating={rating}/>
    </SvgFrame>
  );
}

function GlassExpression({ rating }: { rating: RatingValue }) {
  const stroke = art("effaff");
  if (rating === 1) return <SoftSadExpression ink={stroke}/>;
  if (rating === 2) return <g fill="none" stroke={stroke} strokeLinecap="round"><path d="M31 49q6-5 12 0m14 0q6-5 12 0" strokeWidth="4"/><path d="M35 74q15-11 30 0" strokeWidth="5"/></g>;
  if (rating === 3) return <g fill={stroke} stroke={stroke}><circle cx="37" cy="49" r="4"/><circle cx="63" cy="49" r="4"/><path d="M38 71h24" fill="none" strokeLinecap="round" strokeWidth="5"/></g>;
  if (rating === 4) return <g fill="none" stroke={stroke} strokeLinecap="round"><path d="M31 48q6 7 12 0m14 0q6 7 12 0" strokeWidth="4"/><path d="M33 65q17 19 34 0" fill={art("5038a8")} fillOpacity=".42" strokeWidth="4"/></g>;
  return <g><Heart cx={37} cy={48} color={art("effaff")} fill={art("effaff")} opacity={.9}/><Heart cx={63} cy={48} color={art("effaff")} fill={art("effaff")} opacity={.9}/><path d="M31 65q19 20 38 0" fill={art("38278c")} fillOpacity=".46" stroke={stroke} strokeLinecap="round" strokeWidth="4"/></g>;
}

function MinimalFace({ className, rating }: { className: string; rating: RatingValue }) {
  return <SvgFrame className={className} label={`minimal line rating ${rating}`}><circle cx="50" cy="50" fill={art("fffdf7")} r="41" stroke={art("174f53")} strokeWidth="4"/><path d="M22 38a31 31 0 0 1 56 0" fill="none" stroke={art("c9a24f")} strokeLinecap="round" strokeWidth="4"/><MinimalExpression rating={rating}/><path d="M27 76a31 31 0 0 0 46 0" fill="none" stroke={art("c9a24f")} strokeLinecap="round" strokeWidth="4"/></SvgFrame>;
}

function MinimalExpression({ rating }: { rating: RatingValue }) {
  const ink = art("174f53");
  if (rating === 1) return <SoftSadExpression eyeRadius={3} ink={ink} strokeWidth={4.5}/>;
  if (rating === 2) return <g fill="none" stroke={ink} strokeLinecap="round"><path d="M31 50q6-5 12 0m14 0q6-5 12 0" strokeWidth="4"/><path d="M36 72q14-10 28 0" strokeWidth="4.5"/></g>;
  if (rating === 3) return <g fill={ink} stroke={ink}><circle cx="37" cy="50" r="3.5"/><circle cx="63" cy="50" r="3.5"/><path d="M39 70h22" fill="none" strokeLinecap="round" strokeWidth="4.5"/></g>;
  if (rating === 4) return <g fill="none" stroke={ink} strokeLinecap="round"><path d="M31 49q6 6 12 0m14 0q6 6 12 0" strokeWidth="4"/><path d="M35 64q15 16 30 0" strokeWidth="4.5"/></g>;
  return <g fill="none" stroke={ink}><Heart cx={37} cy={49} color={ink}/><Heart cx={63} cy={49} color={ink}/><path d="M33 65q17 18 34 0" strokeLinecap="round" strokeWidth="4.5"/></g>;
}

function RetroFace({ className, id, rating }: { className: string; id: string; rating: RatingValue }) {
  const fills: Record<RatingValue, string> = { 1: art("f36d5d"), 2: art("f8944e"), 3: art("ffd34f"), 4: art("78cf94"), 5: art("ffd84d") };
  return (
    <SvgFrame className={className} label={`retro sticker rating ${rating}`}>
      <defs><pattern height="6" id={`dots-${id}`} patternUnits="userSpaceOnUse" width="6"><circle cx="1.5" cy="1.5" fill={art("111827")} opacity=".09" r="1"/></pattern></defs>
      <path d="M50 5c25 0 45 20 45 45S75 95 50 95 5 75 5 50 25 5 50 5Z" fill={art("fff")} stroke={art("111827")} strokeWidth="3"/><circle cx="50" cy="50" fill={fills[rating]} r="38" stroke={art("111827")} strokeWidth="5"/><circle cx="50" cy="50" fill={`url(#dots-${id})`} r="36"/><RetroExpression rating={rating}/><path d="M20 76c12 13 46 20 63-5" fill="none" stroke={art("48a9c5")} strokeLinecap="round" strokeWidth="4"/>
    </SvgFrame>
  );
}

function RetroExpression({ rating }: { rating: RatingValue }) {
  const ink = art("111827");
  if (rating === 1) return <SoftSadExpression eyeRadius={4} ink={ink} strokeWidth={5.5}/>;
  if (rating === 2) return <g fill="none" stroke={ink} strokeLinecap="round"><path d="M29 51q8-7 15 0m12 0q7-7 15 0" strokeWidth="5"/><path d="M34 75q16-14 32 0" strokeWidth="6"/></g>;
  if (rating === 3) return <g fill={ink} stroke={ink}><circle cx="37" cy="50" r="4.5"/><circle cx="63" cy="50" r="4.5"/><path d="M36 72h28" fill="none" strokeLinecap="round" strokeWidth="6"/></g>;
  if (rating === 4) return <g fill="none" stroke={ink} strokeLinecap="round"><path d="M29 49q8 8 15 0m12 0q7 8 15 0" strokeWidth="5"/><path d="M31 64q19 22 38 0" fill={art("fff")} strokeWidth="5"/></g>;
  return <g><Star cx={36} cy={47}/><Star cx={64} cy={47}/><circle cx="30" cy="64" fill={art("f46475")} r="5"/><circle cx="70" cy="64" fill={art("f46475")} r="5"/><path d="M30 64q20 26 40 0" fill={art("fff")} stroke={ink} strokeLinecap="round" strokeWidth="5"/><path d="M40 77q10 8 20 0" fill={art("f46475")}/></g>;
}

function SoftSadExpression({ eyeRadius = 3.5, ink, strokeWidth = 5 }: { eyeRadius?: number; ink: string; strokeWidth?: number }) {
  return (
    <g fill="none" stroke={ink} strokeLinecap="round">
      <path d="M30 44l12-2m28 2-12-2" strokeWidth={strokeWidth - 1}/>
      <circle cx="37" cy="52" fill={ink} r={eyeRadius} stroke="none"/>
      <circle cx="63" cy="52" fill={ink} r={eyeRadius} stroke="none"/>
      <path d="M35 74q15-12 30 0" strokeWidth={strokeWidth}/>
    </g>
  );
}

function Heart({ color, cx, cy, fill = "none", opacity = 1 }: { color: string; cx: number; cy: number; fill?: string; opacity?: number }) {
  const centeredCx = cx - 2;
  return <path d={`M${centeredCx} ${cy + 9}c-2-2-10-7-10-14 0-4 3-7 7-7 3 0 5 1 7 4 2-3 4-4 7-4 4 0 7 3 7 7 0 7-8 12-10 14l-4 3-4-3Z`} fill={fill} opacity={opacity} stroke={color} strokeLinejoin="round" strokeWidth="3"/>;
}

function Star({ cx, cy }: { cx: number; cy: number }) {
  const points = [[0,-12],[4,-4],[12,-3],[6,3],[8,11],[0,7],[-8,11],[-6,3],[-12,-3],[-4,-4]].map(([x,y]) => `${cx+x},${cy+y}`).join(" ");
  return <polygon fill={art("fff5a8")} points={points} stroke={art("111827")} strokeLinejoin="round" strokeWidth="4"/>;
}

function IphoneAssetFace({ className, rating }: { className: string; rating: RatingValue }) {
  return (
    <SvgFrame className={className} label={`iphone emoji rating ${rating}`}>
      <image
        height="100"
        href={`/emoji/ios/${rating}.svg`}
        preserveAspectRatio="xMidYMid meet"
        width="100"
        x="0"
        y="0"
      />
    </SvgFrame>
  );
}

function IphoneFace({ className, id, rating }: { className: string; id: string; rating: RatingValue }) {
  const isRed = rating === 1;
  return (
    <SvgFrame className={className} label={`iphone emoji rating ${rating}`}>
      <defs>
        <radialGradient id={`ip-yel-${id}`} cx="36%" cy="28%" r="70%">
          <stop offset="0%" stopColor={art("fff68e")}/>
          <stop offset="24%" stopColor={art("ffd43a")}/>
          <stop offset="68%" stopColor={art("ffa200")}/>
          <stop offset="92%" stopColor={art("ee7000")}/>
          <stop offset="100%" stopColor={art("c54c00")}/>
        </radialGradient>
        <radialGradient id={`ip-red-${id}`} cx="36%" cy="28%" r="70%">
          <stop offset="0%" stopColor={art("ff9668")}/>
          <stop offset="22%" stopColor={art("ff5232")}/>
          <stop offset="65%" stopColor={art("df1616")}/>
          <stop offset="90%" stopColor={art("b00815")}/>
          <stop offset="100%" stopColor={art("73000c")}/>
        </radialGradient>
        <radialGradient id={`ip-glow-yel-${id}`} cx="38%" cy="24%" r="48%">
          <stop offset="0%" stopColor={art("ffffff")} stopOpacity="0.45"/>
          <stop offset="100%" stopColor={art("ffffff")} stopOpacity="0"/>
        </radialGradient>
        <radialGradient id={`ip-glow-red-${id}`} cx="38%" cy="24%" r="48%">
          <stop offset="0%" stopColor={art("ffffff")} stopOpacity="0.38"/>
          <stop offset="100%" stopColor={art("ffffff")} stopOpacity="0"/>
        </radialGradient>
        <linearGradient id={`ip-heart-grad-${id}`} x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" stopColor={art("ff3a58")}/>
          <stop offset="100%" stopColor={art("d60e2e")}/>
        </linearGradient>
        <clipPath id={`ip-mouth-clip-${id}`}>
          <path d="M 31 60 Q 50 62 69 60 C 69 77 61 82 50 82 C 39 82 31 77 31 60 Z"/>
        </clipPath>
        <filter id={`ip-sh-${id}`} x="-15%" y="-10%" width="130%" height="130%">
          <feDropShadow dx="0" dy="3" floodColor={art("000000")} floodOpacity="0.16" stdDeviation="2.5"/>
        </filter>
      </defs>
      <ellipse cx="50" cy="96" rx="30" ry="5.5" fill={art("19383a")} opacity="0.14"/>
      <circle
        cx="50"
        cy="49"
        r="42"
        fill={`url(#${isRed ? `ip-red-${id}` : `ip-yel-${id}`})`}
        filter={`url(#ip-sh-${id})`}
      />
      <ellipse
        cx="44"
        cy="27"
        rx="23"
        ry="13"
        fill={`url(#${isRed ? `ip-glow-red-${id}` : `ip-glow-yel-${id}`})`}
      />
      <IphoneExpression id={id} rating={rating}/>
    </SvgFrame>
  );
}

function IphoneExpression({ id, rating }: { id: string; rating: RatingValue }) {
  if (rating === 1) {
    const ink = art("250407");
    return (
      <g>
        <path d="M 23 35 Q 34 40 44 45" stroke={ink} strokeWidth="4.8" strokeLinecap="round" fill="none"/>
        <path d="M 77 35 Q 66 40 56 45" stroke={ink} strokeWidth="4.8" strokeLinecap="round" fill="none"/>
        <path d="M 48 39 Q 50 36 52 39" stroke={art("5a080d")} strokeWidth="2" strokeLinecap="round" fill="none"/>
        <ellipse cx="35" cy="52" rx="4.2" ry="4.5" fill={ink}/>
        <ellipse cx="65" cy="52" rx="4.2" ry="4.5" fill={ink}/>
        <circle cx="36.5" cy="50.5" r="1.3" fill={art("ffffff")} opacity="0.8"/>
        <circle cx="66.5" cy="50.5" r="1.3" fill={art("ffffff")} opacity="0.8"/>
        <path d="M 33 74 Q 50 61 67 74" stroke={ink} strokeWidth="4.5" strokeLinecap="round" fill="none"/>
      </g>
    );
  }

  const ink = art("3d1e06");

  if (rating === 2) {
    return (
      <g>
        <path d="M 27 38 Q 35 31 44 33" stroke={ink} strokeWidth="3.4" strokeLinecap="round" fill="none"/>
        <path d="M 73 38 Q 65 31 56 33" stroke={ink} strokeWidth="3.4" strokeLinecap="round" fill="none"/>
        <circle cx="35" cy="48" r="4.5" fill={ink}/>
        <circle cx="65" cy="48" r="4.5" fill={ink}/>
        <circle cx="36.5" cy="46.5" r="1.3" fill={art("ffffff")} opacity="0.8"/>
        <circle cx="66.5" cy="46.5" r="1.3" fill={art("ffffff")} opacity="0.8"/>
        <path d="M 34 71 Q 50 59 66 71" stroke={ink} strokeWidth="4.2" strokeLinecap="round" fill="none"/>
      </g>
    );
  }

  if (rating === 3) {
    return (
      <g>
        <circle cx="35" cy="48" r="4.5" fill={ink}/>
        <circle cx="65" cy="48" r="4.5" fill={ink}/>
        <circle cx="36.5" cy="46.5" r="1.3" fill={art("ffffff")} opacity="0.8"/>
        <circle cx="66.5" cy="46.5" r="1.3" fill={art("ffffff")} opacity="0.8"/>
        <path d="M 34 68 H 66" stroke={ink} strokeWidth="4.2" strokeLinecap="round" fill="none"/>
      </g>
    );
  }

  if (rating === 4) {
    return (
      <g>
        <circle cx="35" cy="47" r="4.5" fill={ink}/>
        <circle cx="65" cy="47" r="4.5" fill={ink}/>
        <circle cx="36.5" cy="45.5" r="1.3" fill={art("ffffff")} opacity="0.8"/>
        <circle cx="66.5" cy="45.5" r="1.3" fill={art("ffffff")} opacity="0.8"/>
        <path d="M 33 63 Q 50 77 67 63" stroke={ink} strokeWidth="4.2" strokeLinecap="round" fill="none"/>
      </g>
    );
  }

  return (
    <g>
      <g transform="translate(34, 43) rotate(-11) scale(0.92)">
        <path d="M 0 4 C -2 1 -11 -4 -11 -10 C -11 -15 -6 -18 0 -13 C 6 -18 11 -15 11 -10 C 11 -4 2 1 0 4 Z" fill={`url(#ip-heart-grad-${id})`}/>
        <ellipse cx="-4" cy="-11" rx="2.6" ry="1.5" fill={art("ffffff")} opacity="0.45" transform="rotate(-30 -4 -11)"/>
      </g>
      <g transform="translate(66, 43) rotate(11) scale(0.92)">
        <path d="M 0 4 C -2 1 -11 -4 -11 -10 C -11 -15 -6 -18 0 -13 C 6 -18 11 -15 11 -10 C 11 -4 2 1 0 4 Z" fill={`url(#ip-heart-grad-${id})`}/>
        <ellipse cx="-4" cy="-11" rx="2.6" ry="1.5" fill={art("ffffff")} opacity="0.45" transform="rotate(-30 -4 -11)"/>
      </g>
      <path d="M 31 60 Q 50 62 69 60 C 69 77 61 82 50 82 C 39 82 31 77 31 60 Z" fill={art("3d1810")}/>
      <g clipPath={`url(#ip-mouth-clip-${id})`}>
        <ellipse cx="50" cy="80" rx="12" ry="7" fill={art("f87185")}/>
      </g>
      <path d="M 31 60 Q 50 62 69 60" stroke={art("3d1810")} strokeWidth="2.5" strokeLinecap="round" fill="none"/>
    </g>
  );
}
