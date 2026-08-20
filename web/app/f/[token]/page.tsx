"use client";

import { useEffect, useState } from "react";
import { FlyerSurvey } from "../../../components/flyer-survey";

export default function FlyerSurveyPage({ params }: { params: Promise<{ token: string }> }) {
  const [token, setToken] = useState("");

  useEffect(() => {
    void params.then((value) => setToken(value.token));
  }, [params]);

  return token ? <FlyerSurvey resolve={{ kind: "token", token }} /> : null;
}