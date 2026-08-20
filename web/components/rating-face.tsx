"use client";

import { useId } from "react";

const faceColors = {
  1: ["#ff7a68", "#e8456b"],
  2: ["#ffb45c", "#ff7859"],
  3: ["#ffd66b", "#f3a94b"],
  4: ["#7ed6b2", "#2fa58f"],
  5: ["#5d918a", "#315759"]
} as const;

export type RatingValue = keyof typeof faceColors;

export function RatingFace({ className = "", rating }: { className?: string; rating: RatingValue }) {
  const gradientId = useId().replaceAll(":", "");
  const [start, end] = faceColors[rating];

  return (
    <svg aria-hidden="true" className={className} focusable="false" viewBox="0 0 100 106" xmlns="http://www.w3.org/2000/svg">
      <defs>
        <linearGradient id={gradientId} x1="18" x2="82" y1="12" y2="92" gradientUnits="userSpaceOnUse">
          <stop stopColor={start} />
          <stop offset="1" stopColor={end} />
        </linearGradient>
      </defs>
      <ellipse cx="50" cy="96" fill="#19383a" opacity="0.12" rx="31" ry="6" />
      <circle cx="50" cy="50" fill={`url(#${gradientId})`} r="43" stroke="#ffffff" strokeWidth="3" />
      <path d="M24 31C31 18 46 12 60 16" fill="none" opacity="0.28" stroke="#ffffff" strokeLinecap="round" strokeWidth="7" />
      {rating === 1 ? <AngryExpression /> : null}
      {rating === 2 ? <UnhappyExpression /> : null}
      {rating === 3 ? <NeutralExpression /> : null}
      {rating === 4 ? <HappyExpression /> : null}
      {rating === 5 ? <LovedExpression /> : null}
    </svg>
  );
}

function AngryExpression() {
  return (
    <g fill="none" stroke="#19383a" strokeLinecap="round" strokeLinejoin="round">
      <path d="m29 38 14 5" strokeWidth="4" />
      <path d="m71 38-14 5" strokeWidth="4" />
      <circle cx="37" cy="50" fill="#19383a" r="3.5" stroke="none" />
      <circle cx="63" cy="50" fill="#19383a" r="3.5" stroke="none" />
      <path d="M34 75Q50 59 66 75" strokeWidth="5" />
    </g>
  );
}

function UnhappyExpression() {
  return (
    <g fill="none" stroke="#19383a" strokeLinecap="round">
      <path d="M31 47q6-5 12 0" strokeWidth="4" />
      <path d="M57 47q6-5 12 0" strokeWidth="4" />
      <path d="M35 73Q50 61 65 73" strokeWidth="5" />
    </g>
  );
}

function NeutralExpression() {
  return (
    <g fill="#19383a" stroke="#19383a" strokeLinecap="round">
      <circle cx="37" cy="48" r="4" stroke="none" />
      <circle cx="63" cy="48" r="4" stroke="none" />
      <path d="M37 70h26" fill="none" strokeWidth="5" />
    </g>
  );
}

function HappyExpression() {
  return (
    <g fill="none" stroke="#19383a" strokeLinecap="round">
      <path d="M31 48q6 6 12 0" strokeWidth="4" />
      <path d="M57 48q6 6 12 0" strokeWidth="4" />
      <path d="M33 65q17 20 34 0" fill="#fff7e8" strokeLinejoin="round" strokeWidth="4" />
    </g>
  );
}

function LovedExpression() {
  return (
    <g strokeLinecap="round" strokeLinejoin="round">
      <path d="M38 58C35.8 56.1 28 50.7 28 44.8c0-3.8 2.8-6.8 6.6-6.8 2.3 0 4.3 1.1 5.4 2.9 1.1-1.8 3.1-2.9 5.4-2.9 3.8 0 6.6 3 6.6 6.8 0 5.9-7.8 11.3-10 13.2l-2 1.8Z" fill="#fff3d7" stroke="#19383a" strokeWidth="2.5" />
      <path d="M60 58C57.8 56.1 50 50.7 50 44.8c0-3.8 2.8-6.8 6.6-6.8 2.3 0 4.3 1.1 5.4 2.9 1.1-1.8 3.1-2.9 5.4-2.9 3.8 0 6.6 3 6.6 6.8 0 5.9-7.8 11.3-10 13.2l-2 1.8Z" fill="#fff3d7" stroke="#19383a" strokeWidth="2.5" />
      <path d="M31 66q19 21 38 0" fill="#fff8ec" stroke="#19383a" strokeWidth="4" />
    </g>
  );
}
