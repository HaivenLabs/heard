"use client";

import { useEffect, useState } from "react";
import { FlyerSurvey } from "../../../components/flyer-survey";

export default function FlyerSurveyPathPage({ params }: { params: Promise<{ path: string[] }> }) {
  const [path, setPath] = useState<string[]>([]);

  useEffect(() => {
    void params.then((value) => setPath(value.path ?? []));
  }, [params]);

  if (path.length < 2) {
    return null;
  }

  const [handle, ...rest] = path;
  return <FlyerSurvey resolve={{ kind: "path", handle, slug: rest.join("/") }} />;
}