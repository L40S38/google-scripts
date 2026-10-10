"""Gemini API(Interactions API)に短いプロンプトを送るサンプル。

Usage:
    GEMINI_API_KEY=... python main.py
"""

from google import genai


def main() -> None:
    client = genai.Client()
    interaction = client.interactions.create(
        model="gemini-3.8-flash",
        input="Explain how AI works in a few words",
    )
    print(interaction.output_text)


if __name__ == "__main__":
    main()
