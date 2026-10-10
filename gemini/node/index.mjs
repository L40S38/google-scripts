// Gemini API(Interactions API)に短いプロンプトを送るサンプル。
//
// Usage:
//   GEMINI_API_KEY=... node index.mjs

import { GoogleGenAI } from "@google/genai";

const ai = new GoogleGenAI({});

const interaction = await ai.interactions.create({
  model: "gemini-3.8-flash",
  input: "Explain how AI works in a few words",
});

console.log(interaction.output_text);
