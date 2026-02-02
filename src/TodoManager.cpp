#include "TodoManager.h"
#include <algorithm>

TodoManager::TodoManager(const std::vector<Task>& tasks) : tasks(tasks) {}

void TodoManager::addTask(const Task& task) {
    tasks.push_back(task);
}

void TodoManager::addTasks(const std::vector<Task>& newTasks) {
    tasks.insert(tasks.end(), newTasks.begin(), newTasks.end());
}

std::vector<Task> TodoManager::filterByDate(const std::string& date) const {
    std::vector<Task> result;
    std::copy_if(tasks.begin(), tasks.end(), std::back_inserter(result),
                 [&date](const Task& task) { return task.getDate() == date; });
    return result;
}

std::vector<Task> TodoManager::filterByStatus(TaskStatus status) const {
    std::vector<Task> result;
    std::copy_if(tasks.begin(), tasks.end(), std::back_inserter(result),
                 [&status](const Task& task) { return task.getStatus() == status; });
    return result;
}

std::vector<Task> TodoManager::search(const std::string& keyword) const {
    std::vector<Task> result;
    std::copy_if(tasks.begin(), tasks.end(), std::back_inserter(result),
                 [&keyword](const Task& task) {
                     std::string desc = task.getDescription();
                     std::string kw = keyword;
                     // Case-insensitive search
                     std::transform(desc.begin(), desc.end(), desc.begin(), ::tolower);
                     std::transform(kw.begin(), kw.end(), kw.begin(), ::tolower);
                     return desc.find(kw) != std::string::npos;
                 });
    return result;
}

int TodoManager::getTotalCount() const {
    return static_cast<int>(tasks.size());
}

int TodoManager::getCountByStatus(TaskStatus status) const {
    return static_cast<int>(std::count_if(tasks.begin(), tasks.end(),
                                          [&status](const Task& task) { return task.getStatus() == status; }));
}

std::map<std::string, int> TodoManager::getTaskCountByDate() const {
    std::map<std::string, int> counts;
    for (const auto& task : tasks) {
        counts[task.getDate()]++;
    }
    return counts;
}

const std::vector<Task>& TodoManager::getAllTasks() const {
    return tasks;
}
