#ifndef TODOMANAGER_H
#define TODOMANAGER_H

#include "Task.h"
#include <vector>
#include <string>
#include <map>

class TodoManager {
private:
    std::vector<Task> tasks;

public:
    TodoManager() = default;
    TodoManager(const std::vector<Task>& tasks);

    void addTask(const Task& task);
    void addTasks(const std::vector<Task>& tasks);

    std::vector<Task> filterByDate(const std::string& date) const;
    std::vector<Task> filterByStatus(TaskStatus status) const;
    std::vector<Task> search(const std::string& keyword) const;

    int getTotalCount() const;
    int getCountByStatus(TaskStatus status) const;
    std::map<std::string, int> getTaskCountByDate() const;

    const std::vector<Task>& getAllTasks() const;
};

#endif // TODOMANAGER_H
