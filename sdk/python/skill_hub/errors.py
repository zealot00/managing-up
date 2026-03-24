class SkillHubError(Exception):
    pass


class APIError(SkillHubError):
    def __init__(self, status_code: int, code: str, message: str):
        self.status_code = status_code
        self.code = code
        self.message = message
        super().__init__(f"{code}: {message}")
