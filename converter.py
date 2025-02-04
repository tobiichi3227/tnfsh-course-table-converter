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
class_index_template = env.get_template("classindex.html")
teacher_template = env.get_template("teacher.html")
teacher_index_template = env.get_template("teacherindex.html")


def convert(xls_content: io.BytesIO):
    classes: dict[ClassID, Class] = {}
    classnum_to_class: dict[int, ClassID] = {}
    teachers: dict[TeacherID, Teacher] = {}
    teachers[TeacherID("empty")] = Teacher(
        TeacherID("empty"),
        "",
        [[defaultdict(list) for _ in range(9)] for _ in range(6)],
    )

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
            classnum_to_class[int(r[1])] = class_id
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

        if not teacher_id:
            teacher_id = TeacherID("empty")
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
            if teacher_id == TeacherID("empty"):
                continue
            generated_html = teacher_template.render(
                teacher=teacher,
                teacher_id=teacher_id,
                classes=classes,
                update_timestamp=update_timestamp,
            )
            zip.writestr(f"T{teacher_id}.HTML", generated_html)

        generated_html = class_index_template.render(
            classnum_to_class=classnum_to_class, update_timestamp=update_timestamp
        )
        zip.writestr("_ClassIndex.html", generated_html)

        subjects = [
            {"chinese": "國文科", "range": {"A"}, "teachers": []},
            {"chinese": "英文科", "range": {"B"}, "teachers": []},
            {"chinese": "數學科", "range": {"C"}, "teachers": []},
            {"chinese": "社會科", "range": {"D", "E", "F"}, "teachers": []},
            {"chinese": "自然科", "range": {"G", "H", "I"}, "teachers": []},
            {"chinese": "藝能科", "range": {"J", "K", "L"}, "teachers": []},
            {"chinese": "外聘教師", "range": {}, "teachers": []},
        ]
        r_s = {
            "A": subjects[0],
            "B": subjects[1],
            "C": subjects[2],
            "D": subjects[3],
            "E": subjects[3],
            "F": subjects[3],
            "G": subjects[4],
            "H": subjects[4],
            "I": subjects[4],
            "J": subjects[5],
            "K": subjects[5],
            "L": subjects[5],
        }
        teachers.pop(TeacherID("empty"))
        for teacher_id in sorted(teachers.keys()):
            if not teacher_id:
                continue
            group = teacher_id[0]

            if group in r_s:
                r_s[group]["teachers"].append(teacher_id)
            else:
                subjects[6]["teachers"].append(teacher_id)

        generated_html = teacher_index_template.render(
            teachers=teachers, subjects=subjects, update_timestamp=update_timestamp
        )
        zip.writestr("_TeachIndex.html", generated_html)

    return output_zip
