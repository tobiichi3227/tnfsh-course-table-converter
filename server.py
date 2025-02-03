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


@app.route("/test-download")
def test_download():
    from io import BytesIO
    import zipfile

    zip_buffer = BytesIO()
    with zipfile.ZipFile(zip_buffer, "w", zipfile.ZIP_DEFLATED) as zip_file:
        zip_file.writestr("test.txt", "Hello, this is a test file!")

    zip_buffer.seek(0)
    return send_file(zip_buffer, download_name="test.zip", as_attachment=True)


app.run()
