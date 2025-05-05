import threading
import webbrowser

from flask import Flask, request, send_file, render_template

import converter

app = Flask(__name__)


@app.route("/", methods=["GET", "POST"])
def index():
    if request.method == "POST":
        xls_file = request.files.get("file")
        if not xls_file:
            return "No file uploaded!", 400

        print(xls_file.filename)
        try:
            output_zip = converter.convert(xls_file.stream)
            output_zip.seek(0)
            return send_file(output_zip, download_name="output.zip", as_attachment=True)
        except Exception as e:
            import traceback

            traceback.print_exception(e)
            return f"Error Occured: {e}"
    return render_template("index.html")

if __name__ == "__main__":
    threading.Thread(target=lambda: webbrowser.open('http://localhost:5000')).run()
    app.run()


