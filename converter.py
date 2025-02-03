import io
import zipfile
import datetime
from collections import defaultdict
from dataclasses import dataclass
from typing import NewType, DefaultDict

import xlrd
import jinja2


ClassID = NewType("ClassID", str)
TeacherID = NewType("TeacherID", str)
CourseName = NewType("CourseName", str)


@dataclass
class Class:
    class_id: ClassID
    class_num: int
    courses: list[list[DefaultDict[CourseName, list[TeacherID]]]]


@dataclass
class Teacher:
    teacher_id: TeacherID
    name: str
    courses: list[list[DefaultDict[CourseName, list[ClassID]]]]


Classes = NewType("Classes", dict[ClassID, Class])
Teachers = NewType("Teachers", dict[TeacherID, Teacher])

env = jinja2.Environment(
    loader=jinja2.FileSystemLoader("templates"),
    extensions=["jinja2.ext.loopcontrols"],
)
class_template = env.get_template("class.html")
teacher_template = env.get_template("teacher.html")


def convert(xls_content: io.BytesIO):
    classes: dict[ClassID, Class] = {}
    teachers: dict[TeacherID, Teacher] = {}
    wb = xlrd.open_workbook_xls(file_contents=xls_content.read())
    ws = wb.sheet_by_name("xls")
    rows = ws.nrows
    update_timestamp = datetime.datetime.now().strftime("%Y/%m/%d %H:%M:%S")

    for row in range(rows):
        if row == 0:
            continue
        r = list(map(str, ws.row_values(rowx=row, start_colx=0, end_colx=None)))
        class_id: ClassID = ClassID(r[0])
        if not r[1].isdigit():
            continue

        if class_id not in classes:
            classes[class_id] = Class(
                class_id,
                int(r[1]),
                [[defaultdict(list) for _ in range(9)] for _ in range(6)],
            )
        c = classes[class_id]

        teacher_id: TeacherID = TeacherID(r[9])
        if teacher_id not in teachers:
            teachers[teacher_id] = Teacher(
                teacher_id,
                r[10],
                [[defaultdict(list) for _ in range(9)] for _ in range(6)],
            )
        t = teachers[teacher_id]

        week = int(r[2])
        th = int(r[3])
        course_name: CourseName = CourseName(r[5])

        if teacher_id:
            c.courses[week][th][course_name].append(teacher_id)
            c.courses[week][th][course_name].sort()

        t.courses[week][th][course_name].append(class_id)
        t.courses[week][th][course_name].sort()

    output_zip = io.BytesIO()
    with zipfile.ZipFile(output_zip, "w", compression=zipfile.ZIP_LZMA) as zip:
        for class_id, ccls in classes.items():
            generated_html = class_template.render(
                ccls=ccls,
                class_id=class_id,
                teachers=teachers,
                update_timestamp=update_timestamp,
            )
            zip.writestr(f"C{class_id}.HTML", generated_html)

        for teacher_id, teacher in teachers.items():
            generated_html = teacher_template.render(
                teacher=teacher,
                teacher_id=teacher_id,
                classes=classes,
                update_timestamp=update_timestamp,
            )
            zip.writestr(f"T{teacher_id}.HTML", generated_html)

    return output_zip
