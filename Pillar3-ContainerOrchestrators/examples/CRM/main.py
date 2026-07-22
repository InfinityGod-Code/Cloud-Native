from fastapi import FastAPI, Request
from fastapi.responses import RedirectResponse
from Students.main import app as student_app
from Library.main import app as library_app

app = FastAPI()


class RedirectStudentPortalException(Exception):
    pass


class RedirectLibraryPortalException(Exception):
    pass


@app.exception_handler(RedirectStudentPortalException)
async def redirect_student(request, exc):
    return RedirectResponse(url="/students")


@app.exception_handler(RedirectLibraryPortalException)
async def redirect_library(request, exc):
    return RedirectResponse(url="/library")


@app.get("/")
def get():
    return {"content": "Hello"}


@app.get("/portal/{portal_id}")
def call_api_gateway(request: Request):
    portal_id = request.path_params["portal_id"]
    if portal_id == "1":
        raise RedirectStudentPortalException()
    elif portal_id == "2":
        raise RedirectLibraryPortalException()
    return {"message": "University ERP Systems"}


app.mount("/students", student_app)
app.mount("/library", library_app)
