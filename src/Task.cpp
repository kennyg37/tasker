#include "Task.h"
#include <iostream>
#include <iomanip>

Task::Task(const std::string& date, int taskNumber, const std::string& description, TaskStatus status)
    : date(date), taskNumber(taskNumber), description(description), status(status) {}

// Getters
std::string Task::getDate() const {
    return date;
}

int Task::getTaskNumber() const {
    return taskNumber;
}

std::string Task::getDescription() const {
    return description;
}

TaskStatus Task::getStatus() const {
    return status;
}

std::string Task::getStatusString() const {
    switch (status) {
        case TaskStatus::Done:
            return "Done";
        case TaskStatus::Postponed:
            return "Postponed";
        case TaskStatus::Pending:
            return "Pending";
        case TaskStatus::Other:
            return "Other";
        default:
            return "Unknown";
    }
}

void Task::display() const {
    std::cout << std::setw(12) << date << " | "
              << taskNumber << ". "
              << std::setw(40) << std::left << description
              << " [" << getStatusString() << "]" << std::endl;
}
