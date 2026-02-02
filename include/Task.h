#ifndef TASK_H
#define TASK_H

#include <string>

enum class TaskStatus {
    Done,
    Postponed,
    Pending,
    Other
};

class Task {
private:
    std::string date;
    int taskNumber;
    std::string description;
    TaskStatus status;

public:
    
    Task(const std::string& date, int taskNumber, const std::string& description, TaskStatus status);

    std::string getDate() const;
    int getTaskNumber() const;
    std::string getDescription() const;
    TaskStatus getStatus() const;
    std::string getStatusString() const;

    void display() const;
};

#endif // TASK_H
