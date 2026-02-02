#include "Task.h"
#include "TodoParser.h"
#include "TodoManager.h"
#include "Display.h"
#include <iostream>
#include <string>
#include <algorithm>

int main(int argc, char* argv[]) {
    if (argc < 2) {
        Display display;
        display.showHelp();
        return 0;
    }

    std::string command = argv[1];

    if (command == "help") {
        Display display;
        display.showHelp();
        return 0;
    }

    TodoParser parser;
    std::vector<Task> tasks = parser.parse();
    
    if (tasks.empty()) {
        std::cerr << "No tasks found. Please check your TODO_FILE_PATH and file format." << std::endl;
        return 1;
    }

    TodoManager manager(tasks);
    Display display;

    if (command == "list") {
        display.showTasks(manager.getAllTasks(), "All Tasks");
    }
    else if (command == "stats") {
        display.showStatistics(manager);
    }
    else if (command == "filter") {
        if (argc < 4) {
            std::cerr << "Usage: tasker filter <date|status> <value>" << std::endl;
            return 1;
        }

        std::string filterType = argv[2];
        std::string filterValue = argv[3];

        if (filterType == "date") {
            auto filtered = manager.filterByDate(filterValue);
            display.showTasks(filtered, "Tasks for " + filterValue);
        }
        else if (filterType == "status") {
            std::string statusLower = filterValue;
            std::transform(statusLower.begin(), statusLower.end(), statusLower.begin(), ::tolower);
            
            TaskStatus status;
            if (statusLower == "done") {
                status = TaskStatus::Done;
            } else if (statusLower == "postponed") {
                status = TaskStatus::Postponed;
            } else if (statusLower == "pending") {
                status = TaskStatus::Pending;
            } else {
                status = TaskStatus::Other;
            }

            auto filtered = manager.filterByStatus(status);
            display.showTasks(filtered, "Tasks with status: " + filterValue);
        }
        else {
            std::cerr << "Invalid filter type. Use 'date' or 'status'." << std::endl;
            return 1;
        }
    }
    else if (command == "search") {
        if (argc < 3) {
            std::cerr << "Usage: tasker search <keyword>" << std::endl;
            return 1;
        }

        std::string keyword = argv[2];
        auto results = manager.search(keyword);
        display.showTasks(results, "Search results for: " + keyword);
    }
    else {
        std::cerr << "Unknown command: " << command << std::endl;
        std::cerr << "Run 'tasker help' for usage information." << std::endl;
        return 1;
    }

    return 0;
}
