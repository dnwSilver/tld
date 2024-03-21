using TUI.Controls.Components;
using TUI.Domain;
using YamlDotNet.Serialization;
using YamlDotNet.Serialization.NamingConventions;

namespace TUI.Providers.Dependencies;

public class DependencyRepository
{
    private DependenciesDto? _dependenciesDto;

    private DependenciesDto DependenciesDto
    {
        get
        {
            if (_dependenciesDto is not null)
            {
                return _dependenciesDto;
            }

            var deserializer = new DeserializerBuilder()
                .WithNamingConvention(UnderscoredNamingConvention.Instance)
                .Build();

            using var sr = new StreamReader("dependencies.yaml");
            _dependenciesDto = deserializer.Deserialize<DependenciesDto>(sr.ReadToEnd());

            return _dependenciesDto;
        }
    }

    public IEnumerable<Dependency> Read(string stackName)
    {
        return DependenciesDto.Stacks
            .Single(stack => stack.Name == stackName)
            .Conventions
            .Select(convention =>
            {
                var brand = new Brand(convention.Name, convention.Icon, convention.Color);
                return new Dependency(convention.Version, brand);
            });
    }
}