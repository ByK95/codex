#include <iostream>
#include <unordered_map>
#include <vector>
#include <string>
#include <random>
#include <numeric>

class MarkovChain {
public:
    void train(const std::vector<std::string>& data) {
        for (size_t i = 0; i < data.size() - 1; ++i) {
            model[data[i]][data[i + 1]]++;
        }
    }

    std::vector<std::string> generate(const std::string& start, int length) {
        std::vector<std::string> result = { start };
        std::string current = start;

        for (int i = 1; i < length; ++i) {
            const auto& transitions = model[current];
            if (transitions.empty()) break;

            current = weighted_choice(transitions);
            result.push_back(current);
        }

        return result;
    }

    void print_model() const {
        for (const auto& from : model) {
            std::cout << from.first << " -> ";
            for (const auto& to : from.second) {
                std::cout << to.first << " (" << to.second << ") ";
            }
            std::cout << '\n';
        }
    }

private:
    std::unordered_map<std::string, std::unordered_map<std::string, int>> model;
    std::random_device rd;
    std::mt19937 gen{rd()};

    std::string weighted_choice(const std::unordered_map<std::string, int>& choices) {
        std::vector<std::string> keys;
        std::vector<int> weights;
        for (const auto& pair : choices) {
            keys.push_back(pair.first);
            weights.push_back(pair.second);
        }

        std::discrete_distribution<> dist(weights.begin(), weights.end());
        return keys[dist(gen)];
    }
};

// Example usage
int main() {
    MarkovChain chain;
    std::vector<std::string> sequence = { "walk", "run", "jump", "walk", "run", "walk", "jump" };
    chain.train(sequence);

    chain.print_model();

    auto result = chain.generate("walk", 5);
    for (const auto& s : result) {
        std::cout << s << " ";
    }
    std::cout << "\n";
    return 0;
}
